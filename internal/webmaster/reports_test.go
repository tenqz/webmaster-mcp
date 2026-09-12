package webmaster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHistoryReportPreservesSourceDatesAndMissingMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/hosts/") {
			_, _ = w.Write([]byte(`{"user_id":1}`))
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/search-queries/all/history") {
			t.Error("wrong history endpoint")
		}
		if r.URL.Query().Get("date_from") != "2026-09-01T00:00:00Z" || r.URL.Query().Get("date_to") != "2026-09-02T00:00:00Z" {
			t.Error("date contract changed")
		}
		_, _ = w.Write([]byte(`{"indicators":{"TOTAL_SHOWS":[{"date":"2026-09-01T00:00:00+03:00","value":10}]},"upstream_extension":"retained"}`))
	}))
	defer ts.Close()
	p, e := NewClient(ts.Client(), "test", ts.URL, Options{}).GetQueryHistory(context.Background(), QueryHistory{HostID: "https:example.invalid:443", DateFrom: "2026-09-01", DateTo: "2026-09-02"})
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(p)
	if !strings.Contains(string(raw), "+03:00") || strings.Contains(string(raw), "TOTAL_CLICKS") || p["upstream_extension"] != "retained" {
		t.Fatal("source semantics lost", string(raw))
	}
}
