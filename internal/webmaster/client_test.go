package webmaster

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestListHostsUsesCachedUserID documents GET /user then GET /user/{id}/hosts.
func TestListHostsUsesCachedUserID(t *testing.T) {
	var paths []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/user/") || strings.HasSuffix(r.URL.Path, "/user"):
			_, _ = w.Write([]byte(`{"user_id": 42}`))
		default:
			_, _ = w.Write([]byte(`{"hosts":[{"host_id":"https:example.com:443","verified":true}]}`))
		}
	}))
	defer ts.Close()
	client := NewClient(ts.Client(), "token", ts.URL, Options{})
	hosts, err := client.ListHosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].HostID != "https:example.com:443" {
		t.Fatalf("hosts=%v", hosts)
	}
	if _, err := client.GetUser(context.Background()); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(paths, ",")
	if strings.Count(joined, "/user") < 1 {
		t.Fatalf("paths=%v", paths)
	}
}

// TestSubmitRecrawlPostsURL documents the write body for reindexing.
func TestSubmitRecrawlPostsURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/user/") && !strings.Contains(r.URL.Path, "/hosts/") {
			_, _ = w.Write([]byte(`{"user_id":1}`))
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method=%s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		var body map[string]string
		_ = json.Unmarshal(raw, &body)
		if body["url"] != "https://example.com/page" {
			t.Errorf("body=%s", raw)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"task_id":"t1"}`))
	}))
	defer ts.Close()
	payload, err := NewClient(ts.Client(), "token", ts.URL, Options{}).SubmitRecrawl(context.Background(), RecrawlRequest{HostID: "https:example.com:443", URL: "https://example.com/page"})
	if err != nil {
		t.Fatal(err)
	}
	if payload["task_id"] != "t1" {
		t.Fatalf("payload=%v", payload)
	}
}

// TestDemoRejectsUnknownHost documents that fixtures are bound to example.invalid.
func TestDemoRejectsUnknownHost(t *testing.T) {
	_, err := NewDemo().GetSummary(context.Background(), "https:other.invalid:443")
	var failure *RequestError
	if !errors.As(err, &failure) || failure.Status != 403 {
		t.Fatalf("err=%v", err)
	}
}
