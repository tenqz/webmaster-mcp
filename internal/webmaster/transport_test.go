package webmaster

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestRetriesReplayReadOnlyPost protects body replay and bounded retries for transient failures.
func TestRetriesReplayReadOnlyPost(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if string(raw) != `{"value":1}` {
			t.Errorf("replayed body=%s", raw)
		}
		if calls.Add(1) < 3 {
			w.WriteHeader(503)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()
	client := NewClient(ts.Client(), "token", ts.URL, Options{MaxAttempts: 3})
	var result map[string]bool
	if err := client.postJSON(context.Background(), ts.URL, map[string]int{"value": 1}, &result); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 || !result["ok"] {
		t.Fatal("retry did not preserve response")
	}
}

// TestPermanentErrorsDoNotRetry protects quotas and avoids returning raw upstream secrets.
func TestPermanentErrorsDoNotRetry(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(403)
		_, _ = w.Write([]byte("secret upstream body"))
	}))
	defer ts.Close()
	err := NewClient(ts.Client(), "token", ts.URL, Options{}).getJSON(context.Background(), ts.URL, nil)
	var failure *RequestError
	if !errors.As(err, &failure) || failure.Status != 403 || failure.Retryable || calls.Load() != 1 || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unexpected error=%v calls=%d", err, calls.Load())
	}
}

// TestRetryAfterBeyondBudgetDoesNotRetry verifies that Retry-After is respected rather than shortened.
func TestRetryAfterBeyondBudgetDoesNotRetry(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(429)
	}))
	defer ts.Close()
	err := NewClient(ts.Client(), "token", ts.URL, Options{RequestTimeout: time.Second}).getJSON(context.Background(), ts.URL, nil)
	var failure *RequestError
	if !errors.As(err, &failure) || failure.RetryAfterSeconds != 60 || calls.Load() != 1 {
		t.Fatalf("error=%v calls=%d", err, calls.Load())
	}
}

// TestOAuthHeaderIsSent documents the Webmaster Authorization scheme.
func TestOAuthHeaderIsSent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "OAuth secret-token" {
			t.Errorf("Authorization=%q", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()
	var result map[string]bool
	if err := NewClient(ts.Client(), "secret-token", ts.URL, Options{}).getJSON(context.Background(), ts.URL, &result); err != nil {
		t.Fatal(err)
	}
}
