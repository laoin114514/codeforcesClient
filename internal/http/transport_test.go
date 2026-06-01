package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(10)
	start := time.Now()
	for range 3 {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < 200*time.Millisecond {
		t.Errorf("too fast: %v (expected >= 200ms for 3 reqs at 10rps)", elapsed)
	}
}

func TestRateLimiterZeroDisables(t *testing.T) {
	rl := NewRateLimiter(0)
	start := time.Now()
	for range 100 {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("should be instant with rate=0, took %v", elapsed)
	}
}

func TestTransportRetry(t *testing.T) {
	retries := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++
		if retries < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := NewTransport(srv.Client(), NewRateLimiter(0))
	req, _ := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if retries != 3 {
		t.Errorf("retries = %d, want 3", retries)
	}
}

func TestRateLimiterContextCancel(t *testing.T) {
	rl := NewRateLimiter(1)
	_ = rl.Wait(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := rl.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
