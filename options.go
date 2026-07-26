package codeforcesClient

import (
	"context"
	"net/http"
	"time"

	internalhttp "github.com/laoin114514/codeforcesClient/internal/http"
)

// ClientOption 是函数式选项，用于配置 Client。
// 通过 NewClient 传入，如 NewClient(WithRateLimit(5), WithSigner(s))。
type ClientOption func(*Client)

// WithHTTPClient 设置自定义 *http.Client。
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithSigner 设置请求签名器，用于认证接口。
func WithSigner(s Signer) ClientOption {
	return func(c *Client) {
		c.signer = s
	}
}

// WithRateLimit 设置每秒最大请求数。rps <= 0 表示不限流。
func WithRateLimit(rps int) ClientOption {
	return func(c *Client) {
		if rps > 0 {
			c.limiter = internalhttp.NewRateLimiter(rps)
		}
	}
}

// WithBaseURL 设置自定义 API 地址，如测试时指向 mock 服务器。
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithContext 设置 Client 的默认 context，用于所有请求。
// 需要单次超时控制时，建议使用 client.WithContext(ctx) 而非此选项。
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.ctx = ctx
	}
}

// defaultClient 返回带有合理默认值的 Client。
// 默认：codeforces.com/api、10s 超时、不限流、无认证。
func defaultClient() *Client {
	hc := &http.Client{Timeout: 10 * time.Second}
	return &Client{
		ctx:        context.Background(),
		httpClient: hc,
		baseURL:    "https://codeforces.com/api/",
		limiter:    internalhttp.NewRateLimiter(0),
	}
}
