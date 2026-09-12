package webmaster

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecrawlTimeoutDoesNotInviteRetry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`{"user_id":1}`))
			return
		}
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"task_id":"accepted"}`))
	}))
	defer ts.Close()
	_, e := NewClient(ts.Client(), "test", ts.URL, Options{RequestTimeout: 100 * time.Millisecond}).SubmitRecrawl(context.Background(), RecrawlRequest{HostID: "https:example.invalid:443", URL: "https://example.invalid/"})
	var failure *RequestError
	if !errors.As(e, &failure) || failure.Kind != "outcome_unknown" || failure.Retryable {
		t.Fatal("ambiguous write marked retryable", e)
	}
}
