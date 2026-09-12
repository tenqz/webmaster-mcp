package webmaster

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRecrawlUnknownOutcomeIsNeverReplayed(t *testing.T) {
	var writes atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`{"user_id":1}`))
			return
		}
		writes.Add(1)
		w.WriteHeader(503)
	}))
	defer ts.Close()
	_, err := NewClient(ts.Client(), "test", ts.URL, Options{MaxAttempts: 3}).SubmitRecrawl(context.Background(), RecrawlRequest{HostID: "https:example.invalid:443", URL: "https://example.invalid/"})
	var failure *RequestError
	if !errors.As(err, &failure) || failure.Kind != "outcome_unknown" || failure.Retryable || writes.Load() != 1 {
		t.Fatalf("err=%v writes=%d", err, writes.Load())
	}
}

func TestDiscoveryWaitHonorsCancellation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_, _ = w.Write([]byte(`{"user_id":1}`))
	}))
	defer ts.Close()
	c := NewClient(ts.Client(), "test", ts.URL, Options{})
	done := make(chan struct{})
	go func() { defer close(done); _, _ = c.GetUser(context.Background()) }()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := make(chan error, 1)
	go func() { _, err := c.GetUser(ctx); result <- err }()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err=%v", err)
		}
	case <-time.After(time.Second):
		t.Error("cancelled caller waited for discovery")
	}
	close(release)
	<-done
}
