package codeforcessdk

import (
	"net/http"
	"time"

	internalhttp "github.com/laoin114514/codeforcesSDK/internal/http"
)

type ClientOption func(*Client)

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

func WithSigner(s Signer) ClientOption {
	return func(c *Client) {
		c.signer = s
	}
}

func WithRateLimit(rps int) ClientOption {
	return func(c *Client) {
		if rps > 0 {
			c.limiter = internalhttp.NewRateLimiter(rps)
		}
	}
}

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

func defaultClient() *Client {
	hc := &http.Client{Timeout: 10 * time.Second}
	return &Client{
		httpClient: hc,
		baseURL:    "https://codeforces.com/api/",
		limiter:    internalhttp.NewRateLimiter(0),
	}
}
