package http

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	rate     int
	interval time.Duration
	mu       sync.Mutex
	last     time.Time
}

func NewRateLimiter(rps int) *RateLimiter {
	r := &RateLimiter{rate: rps}
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
	return r
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.rate <= 0 {
		return nil
	}
	r.mu.Lock()
	elapsed := time.Since(r.last)
	if elapsed < r.interval {
		r.mu.Unlock()
		select {
		case <-time.After(r.interval - elapsed):
		case <-ctx.Done():
			return ctx.Err()
		}
		r.mu.Lock()
	}
	r.last = time.Now()
	r.mu.Unlock()
	return nil
}

func (r *RateLimiter) SetRate(rps int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rate = rps
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
}

type Transport struct {
	client     *http.Client
	limiter    *RateLimiter
	maxRetries int
	retryWait  time.Duration
}

func NewTransport(client *http.Client, limiter *RateLimiter) *Transport {
	return &Transport{
		client:     client,
		limiter:    limiter,
		maxRetries: 3,
		retryWait:  500 * time.Millisecond,
	}
}

func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i <= t.maxRetries; i++ {
		if err := t.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		resp, err := t.client.Do(req)
		if err != nil {
			lastErr = err
			if i < t.maxRetries {
				t.backoff(ctx, i)
				continue
			}
			return nil, lastErr
		}
		if t.isRetryable(resp.StatusCode) && i < t.maxRetries {
			resp.Body.Close()
			t.backoff(ctx, i)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func (t *Transport) isRetryable(code int) bool {
	return code == 429 || code == 503 || code >= 500
}

func (t *Transport) backoff(ctx context.Context, attempt int) {
	wait := t.retryWait * (1 << attempt)
	select {
	case <-time.After(wait):
	case <-ctx.Done():
	}
}
