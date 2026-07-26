// Package http 提供带限流和重试功能的 HTTP 传输层。
package http

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// RateLimiter 控制请求速率，保证并发安全。
// rps <= 0 表示不限流。
type RateLimiter struct {
	rate     int
	interval time.Duration // 两次请求的最小间隔
	mu       sync.Mutex
	last     time.Time
}

// NewRateLimiter 创建限流器，rps 为每秒允许的请求数。
// rps <= 0 时限流禁用，Wait 直接返回。
func NewRateLimiter(rps int) *RateLimiter {
	r := &RateLimiter{rate: rps}
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
	return r
}

// Wait 阻塞直到可以发送请求而不超过速率限制。
// context 取消时返回 ctx.Err()。
func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.rate <= 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		elapsed := time.Since(r.last)
		if elapsed >= r.interval {
			r.last = time.Now()
			return nil
		}
		waitTime := r.interval - elapsed
		r.mu.Unlock()
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			r.mu.Lock()
			return ctx.Err()
		}
		r.mu.Lock()
		// 重新检查：睡眠期间可能有其他协程抢先通过。
	}
}

// SetRate 运行时动态调整限流速率。
func (r *RateLimiter) SetRate(rps int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rate = rps
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
}

// Transport 在 http.Client 基础上叠加限流和重试。
// 对 429、503、5xx 状态码最多重试 3 次，采用指数退避（500ms/1s/2s）。
type Transport struct {
	client     *http.Client
	limiter    *RateLimiter
	maxRetries int
	retryWait  time.Duration
}

// NewTransport 创建传输层。默认 maxRetries=3, retryWait=500ms。
func NewTransport(client *http.Client, limiter *RateLimiter) *Transport {
	return &Transport{
		client:     client,
		limiter:    limiter,
		maxRetries: 3,
		retryWait:  500 * time.Millisecond,
	}
}

// Do 发送请求，包含限流等待和自动重试。
// 成功时返回响应，重试耗尽后返回最后一次错误。
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

// isRetryable 判断 HTTP 状态码是否值得重试。
func (t *Transport) isRetryable(code int) bool {
	return code == 429 || code == 503 || code >= 500
}

// backoff 等待 retryWait * 2^attempt，或直到 ctx 取消。
func (t *Transport) backoff(ctx context.Context, attempt int) {
	wait := t.retryWait * (1 << attempt)
	select {
	case <-time.After(wait):
	case <-ctx.Done():
	}
}
