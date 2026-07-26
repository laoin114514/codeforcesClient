package codeforcesClient

import "fmt"

// ErrorCode 分类标识 API 返回的错误类型。
type ErrorCode int

const (
	ErrNetwork     ErrorCode = iota // 网络或 HTTP 传输错误
	ErrAPI                          // Codeforces API 返回 status "FAILED"
	ErrRateLimit                    // HTTP 429 触发限流
	ErrAuth                         // 认证或签名失败
	ErrInvalidParam                 // 请求参数无效
)

// CFError 是所有 Client 方法返回的错误类型。
// 使用 errors.As(err, &cfErr) 提取后，通过 Code 字段区分错误类型。
type CFError struct {
	Code    ErrorCode // 错误分类
	Message string    // 人类可读的描述
	Cause   error     // 底层错误，支持 errors.Unwrap
}

// Error 实现 error 接口。
func (e *CFError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回底层错误，支持 errors.Is / errors.As。
func (e *CFError) Unwrap() error {
	return e.Cause
}

var _ error = (*CFError)(nil)
