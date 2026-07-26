package codeforcesClient

import (
	"context"
	"fmt"
	"time"

	"github.com/laoin114514/codeforcesSDK/internal/params"
	"github.com/laoin114514/codeforcesSDK/internal/signature"
)

type Signer interface {
	Sign(ctx context.Context, method string, m map[string]any) (*SignedRequest, error)
}

type SignedRequest struct {
	URL string
}

type StaticSigner struct {
	apiKey string
	secret string
}

func NewStaticSigner(apiKey, secret string) *StaticSigner {
	return &StaticSigner{apiKey: apiKey, secret: secret}
}

func (s *StaticSigner) Sign(_ context.Context, method string, m map[string]any) (*SignedRequest, error) {
	if s.apiKey == "" || s.secret == "" {
		return nil, &CFError{Code: ErrAuth, Message: "apiKey and secret are required"}
	}
	return buildSignedURL(method, s.apiKey, s.secret, m), nil
}

type PoolSigner struct {
	keys map[string]struct{ apiKey, secret string }
}

func NewPoolSigner(keys map[string]struct{ ApiKey, Secret string }) *PoolSigner {
	pool := make(map[string]struct{ apiKey, secret string }, len(keys))
	for h, k := range keys {
		pool[h] = struct{ apiKey, secret string }{k.ApiKey, k.Secret}
	}
	return &PoolSigner{keys: pool}
}

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

func WithHandle(ctx context.Context, handle string) context.Context {
	return context.WithValue(ctx, handleKey, handle)
}

func HandleFromContext(ctx context.Context) (string, bool) {
	h, ok := ctx.Value(handleKey).(string)
	return h, ok
}

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
