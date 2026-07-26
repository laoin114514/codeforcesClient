# Codeforces SDK for Go

[Codeforces API](https://codeforces.com/apiHelp) 的 Go SDK，覆盖全部 16 个官方接口。

## 安装

```bash
go get github.com/laoin114514/codeforcesClient
```

## 快速开始

```go
package main

import (
    "fmt"
    cf "github.com/laoin114514/codeforcesClient"
)

func main() {
    client := cf.NewClient()
    resp, err := client.UserInfo(&cf.UserInfoParams{
        Handles: "tourist;Petr",
    })
    if err != nil {
        panic(err)
    }
    for _, user := range resp.Result {
        fmt.Printf("%s: rating=%d\n", user.Handle, user.Rating)
    }
}
```

## 认证

需要授权的接口（如 `UserFriends`）通过 `Signer` 传入 API Key：

```go
client := cf.NewClient(cf.WithSigner(cf.NewStaticSigner("your-api-key", "your-secret")))
```

多用户场景使用 `PoolSigner`，通过 context 传递 handle：

```go
signer := cf.NewPoolSigner(map[string]struct{ ApiKey, Secret string }{
    "user1": {ApiKey: "...", Secret: "..."},
    "user2": {ApiKey: "...", Secret: "..."},
})
client := cf.NewClient(cf.WithSigner(signer))
ctx := cf.WithHandle(context.Background(), "user1")
resp, err := client.WithContext(ctx).UserFriends(&cf.UserFriendsParams{})
```

## 配置选项

| 选项 | 说明 |
|--------|-------------|
| `WithHTTPClient(client)` | 自定义 `*http.Client` |
| `WithSigner(signer)` | API Key 签名器 |
| `WithRateLimit(rps)` | 每秒最大请求数 |
| `WithBaseURL(url)` | 自定义 API 地址 |

## 错误处理

使用 `errors.As` 区分错误类型：

```go
resp, err := client.UserStatus(params)
var cfErr *cf.CFError
if errors.As(err, &cfErr) {
    switch cfErr.Code {
    case cf.ErrAPI:
        // Codeforces 返回 FAILED
    case cf.ErrRateLimit:
        // 触发频率限制
    case cf.ErrAuth:
        // 认证失败
    }
}
```

## 原始请求

对于未封装的自定义请求：

```go
body, err := client.RawRequest("user.status", &cf.UserStatusParams{
    Handle: "tourist",
})
```

## API 方法

| 分类 | 方法 |
|----------|---------|
| BlogEntry | `BlogEntryComments`, `BlogEntryView` |
| Contest | `ContestHacks`, `ContestList`, `ContestRatingChanges`, `ContestStandings`, `ContestStatus` |
| ProblemSet | `ProblemsetProblems`, `ProblemsetRecentStatus` |
| User | `UserBlogEntries`, `UserFriends`, `UserInfo`, `UserRatedList`, `UserRating`, `UserStatus`, `UserRecentActions` |
