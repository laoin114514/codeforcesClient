package codeforcesClient

import (
	"context"
	"fmt"
	"time"

	"github.com/laoin114514/codeforcesSDK/internal/params"
	"github.com/laoin114514/codeforcesSDK/internal/signature"
)

// Signer 对 API 请求进行签名，注入 apiKey、time 和 apiSig 参数。
// 实现：StaticSigner（单用户）、PoolSigner（多用户）。
type Signer interface {
	Sign(ctx context.Context, method string, m map[string]any) (*SignedRequest, error)
}

// SignedRequest 包含签名后的 URL 路径和查询串。
type SignedRequest struct {
	URL string
}

// StaticSigner 使用固定的 API Key 和 Secret 进行签名。
// 适用于单用户场景。
type StaticSigner struct {
	apiKey string
	secret string
}

// NewStaticSigner 为单个用户创建签名器。
func NewStaticSigner(apiKey, secret string) *StaticSigner {
	return &StaticSigner{apiKey: apiKey, secret: secret}
}

// Sign 实现 Signer 接口。apiKey 或 secret 为空时返回 ErrAuth。
func (s *StaticSigner) Sign(_ context.Context, method string, m map[string]any) (*SignedRequest, error) {
	if s.apiKey == "" || s.secret == "" {
		return nil, &CFError{Code: ErrAuth, Message: "apiKey and secret are required"}
	}
	return buildSignedURL(method, s.apiKey, s.secret, m), nil
}

// PoolSigner 管理多个 handle 到 API Key 的映射。
// 目标 handle 通过 context 传递（WithHandle / HandleFromContext）。
// 适用于需要代表多个用户调用 API 的场景。
type PoolSigner struct {
	keys map[string]struct{ apiKey, secret string }
}

// NewPoolSigner 从 handle→凭据 的映射创建 PoolSigner。
func NewPoolSigner(keys map[string]struct{ ApiKey, Secret string }) *PoolSigner {
	pool := make(map[string]struct{ apiKey, secret string }, len(keys))
	for h, k := range keys {
		pool[h] = struct{ apiKey, secret string }{k.ApiKey, k.Secret}
	}
	return &PoolSigner{keys: pool}
}

// Sign 实现 Signer 接口。从 context 中读取 handle 并查找对应凭据。
// handle 缺失或未注册时返回 ErrAuth。
func (s *PoolSigner) Sign(ctx context.Context, method string, m map[string]any) (*SignedRequest, error) {
	handle, ok := HandleFromContext(ctx)
	if !ok || handle == "" {
		return nil, &CFError{Code: ErrAuth, Message: "handle not found in context"}
	}
	key, ok := s.keys[handle]
	if !ok {
		return nil, &CFError{Code: ErrAuth, Message: fmt.Sprintf("no apiKey for handle %q", handle)}
	}
	return buildSignedURL(method, key.apiKey, key.secret, m), nil
}

type ctxKey struct{}

var handleKey ctxKey

// WithHandle 将 handle 存入 context，供 PoolSigner 查找凭据。
func WithHandle(ctx context.Context, handle string) context.Context {
	return context.WithValue(ctx, handleKey, handle)
}

// HandleFromContext 从 context 中取出 WithHandle 存入的 handle。
func HandleFromContext(ctx context.Context) (string, bool) {
	h, ok := ctx.Value(handleKey).(string)
	return h, ok
}

// buildSignedURL 构造 Codeforces 兼容的签名 URL。
// 注入 apiKey、time，并计算 apiSig=randomPrefix+SHA512(sigInput)。
func buildSignedURL(method, apiKey, secret string, m map[string]any) *SignedRequest {
	allParams := make(map[string]any, len(m)+2)
	for k, v := range m {
		allParams[k] = v
	}
	allParams["apiKey"] = apiKey
	allParams["time"] = time.Now().Unix()

	ordered := params.ToOrderedString(allParams)
	randomPrefix := signature.RandomPrefix()
	sigInput := fmt.Sprintf("%s/%s?%s#%s", randomPrefix, method, ordered, secret)
	hash := signature.SHA512Sum(sigInput)

	return &SignedRequest{
		URL: fmt.Sprintf("%s?%s&apiSig=%s%s", method, ordered, randomPrefix, hash),
	}
}
