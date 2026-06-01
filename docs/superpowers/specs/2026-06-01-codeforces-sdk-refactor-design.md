# Codeforces SDK 重构设计

## 概述

将现有 Codeforces API SDK 从一个仅做 URL 拼接的工具重构为完整的 Go HTTP SDK，支持类型化的 API 请求/响应、签名认证、限流重试、错误分类等工程化能力。

## 目标

- 覆盖 Codeforces 全部 16 个 API 方法，提供类型安全的请求和响应
- 工程化：完整的错误处理、HTTP 传输层、限流、重试、签名机制
- 易于使用：单个 import path，函数式选项模式，开箱即用的默认配置
- 可测试：单元测试（mock）+ 集成测试（build tag）
- 可发布：符合 Go 社区规范的开源库

## 技术栈

- Go 1.24.x，使用泛型等新特性
- 零外部依赖（仅标准库）
- Module: `github.com/laoin114514/codeforcesSDK`

## 项目结构

```
codeforcesSDK/
├── client.go              // Client 结构体, NewClient(), 全部 16 个 API 方法
├── models.go              // 所有 Request/Response 类型定义
├── errors.go              // CFError 类型, 预定义错误常量和错误码
├── options.go             // ClientOption 函数式选项模式
├── signer.go              // Signer 接口, StaticSigner, PoolSigner
├── codeforces_test.go     // 集成测试 (build tag: integration)
├── go.mod
├── go.sum
├── README.md
└── internal/
    ├── http/
    │   ├── transport.go   // HTTP 传输层 (超时、重试、限流)
    │   └── transport_test.go
    ├── signature/
    │   ├── hash.go        // SHA512 签名逻辑
    │   └── hash_test.go
    └── params/
        ├── encoder.go     // 结构体 → 有序参数字符串
        └── encoder_test.go
```

**原则**: 根包暴露所有公开 API 类型，`internal/` 隐藏实现细节。用户只需 `import "github.com/laoin114514/codeforcesSDK"`。

---

## 核心类型设计

### Client

```go
type Client struct {
    httpClient *http.Client
    signer     Signer
    limiter    *internalhttp.RateLimiter
    baseURL    string
}
```

- 通过 `NewClient(opts ...ClientOption)` 创建
- 选项: `WithHTTPClient()`, `WithSigner()`, `WithRateLimit(rps int)`, `WithBaseURL(url string)`
- 默认值: 无签名（仅公开接口可用）、10s 超时、无限流、`https://codeforces.com/api/`

### Signer 接口

```go
type Signer interface {
    Sign(method string, params map[string]any) (*SignedRequest, error)
}

type SignedRequest struct {
    URL string
}
```

- `StaticSigner`: 持有一组 apiKey/secret，所有请求共用
- `PoolSigner`: 持有 `map[handle]{apiKey, secret}`，通过 context 传入 handle 选择 key
- Sign 内部逻辑: 参数排序 → 拼接 → SHA512 签名 → 追加 apiSig + random → 返回完整 URL

### CFError

```go
type CFError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

type ErrorCode int

const (
    ErrNetwork     ErrorCode = iota // 网络错误
    ErrAPI                          // Codeforces API 返回 FAILED
    ErrRateLimit                    // 限流被拒
    ErrAuth                         // 签名/认证失败
    ErrInvalidParam                 // 参数校验失败
)
```

- 实现 `error`, `Unwrap()`, `Is()` 接口
- 预定义常用错误实例
- HTTP 传输层在遇到各类错误时自动包装为对应的 CFError

---

## HTTP 传输层 (`internal/http/`)

### Transport

```go
type Transport struct {
    client     *http.Client
    limiter    *RateLimiter
    maxRetries int
    retryWait  time.Duration
}

func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error)
```

- 请求前: `limiter.Wait(ctx)` 等待令牌
- 遇到 429/503/5xx 自动重试，指数退避 (500ms, 1s, 2s)，最多 `maxRetries` 次（默认 3）
- 返回的 error 被包装为 `CFError{Code: ErrNetwork}`
- 尊重 ctx 取消

### RateLimiter

```go
type RateLimiter struct {
    rate     int
    interval time.Duration
    mu       sync.Mutex
    lastTime time.Time
}
```

- 简单的间隔控制实现，无外部依赖
- `SetRate(rps int)`: 动态调整每秒最大请求数
- `Wait(ctx)`: 达到上限时阻塞，ctx 取消时立即返回 `context.Canceled`

---

## API 方法设计

### 方法签名模式

每个 API 方法同时接收 context 和参数，返回类型化响应 + error:

```go
func (c *Client) UserStatus(ctx context.Context, params *UserStatusParams) (*UserStatusResponse, error)
```

### 覆盖的 16 个方法

| 分类 | 方法 |
|------|------|
| Blog Entry | `BlogEntryComments`, `BlogEntryView` |
| Contest | `ContestHacks`, `ContestList`, `ContestRatingChanges`, `ContestStandings`, `ContestStatus` |
| ProblemSet | `ProblemsetProblems`, `ProblemsetRecentStatus` |
| User | `UserBlogEntries`, `UserFriends`, `UserInfo`, `UserRatedList`, `UserRating`, `UserStatus`, `UserRecentActions` |

### Request/Response 定义

- Params 结构体使用 JSON tag + omitempty，序列化为有序参数字符串
- Response 遵循 Codeforces 格式: `{status: "OK"|"FAILED", comment?: string, result: T}`
- Result 中的业务类型参考 [Codeforces API 文档](https://codeforces.com/apiHelp/objects) 定义
- 所有 models 类型放在 `models.go`

### Raw 兜底

```go
func (c *Client) RawRequest(ctx context.Context, method string, params any) ([]byte, error)
```

- 返回原始 JSON body，不进行反序列化
- 适用于未覆盖的边缘场景或自定义解析需求

---

## 签名逻辑 (`internal/signature/`)

```go
func SignURL(baseURL, method, apiKey, secret, randomPrefix string, params map[string]any) (string, error)
```

- 参数按 key 字母序排列拼接
- 计算 `SHA512(randomPrefix/method?sortedParams#secret)`
- 拼接最终 URL: `baseURL + method?params&apiSig=randomPrefix+hash`

---

## 参数编码 (`internal/params/`)

```go
func Encode(v any, extra map[string]any) (map[string]any, error)
func ToOrderedString(m map[string]any) string
```

- 将 struct 通过 JSON marshal/unmarshal 转为 `map[string]any`
- 按 key 排序拼接为 `key1=val1&key2=val2` 格式
- 处理 `omitempty` 零值过滤
- 支持动态追加参数（如 apiKey, time）

---

## 测试策略

### 单元测试

- `internal/http/transport_test.go`: mock HTTP server 验证重试、退避、限流行为
- `internal/signature/hash_test.go`: 用已知输入/输出验证 SHA512 签名正确性
- `internal/params/encoder_test.go`: 验证参数序列化、排序、omitempty 行为
- 根包 mock `Signer` 和 HTTP transport 验证 Client 方法逻辑

### 集成测试

- 文件: `codeforces_test.go`, build tag: `integration`
- 连接真实 Codeforces API 验证端到端行为
- 需要环境变量 `CF_API_KEY` 和 `CF_SECRET` (可选，仅测试授权接口)
- 运行: `go test -tags=integration -count=1 ./...`

---

## 迁移策略

现有代码将被重构而非渐进式修改：
- `generator.go` 的逻辑拆分到 `client.go` + `internal/signature/` + `internal/params/`
- `struct_transfer.go` 重写为 `internal/params/encoder.go`，修复 nil 返回的静默错误
- `hashEncode.go` 迁移到 `internal/signature/hash.go`
- `params.go` 的 Request 类型迁入 `models.go`，补充 Response 类型
- 删除旧的子结构体模式 (`user{}`, `contest{}`, `problemSet{}`)
