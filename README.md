# Codeforces SDK for Go

Go SDK for the [Codeforces API](https://codeforces.com/apiHelp), covering all 16 official methods.

## Installation

```bash
go get github.com/laoin114514/codeforcesSDK
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    cf "github.com/laoin114514/codeforcesSDK"
)

func main() {
    client := cf.NewClient()
    resp, err := client.UserInfo(context.Background(), &cf.UserInfoParams{
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

## Authentication

For authorized methods (like `UserFriends`), use a `Signer`:

```go
client := cf.NewClient(cf.WithSigner(cf.NewStaticSigner("your-api-key", "your-secret")))
```

For multi-user scenarios, use `PoolSigner` with context-based handle resolution:

```go
signer := cf.NewPoolSigner(map[string]struct{ ApiKey, Secret string }{
    "user1": {ApiKey: "...", Secret: "..."},
    "user2": {ApiKey: "...", Secret: "..."},
})
client := cf.NewClient(cf.WithSigner(signer))
ctx := cf.WithHandle(context.Background(), "user1")
resp, err := client.UserFriends(ctx, &cf.UserFriendsParams{})
```

## Options

| Option | Description |
|--------|-------------|
| `WithHTTPClient(client)` | Custom `*http.Client` |
| `WithSigner(signer)` | API key signer |
| `WithRateLimit(rps)` | Max requests per second |
| `WithBaseURL(url)` | Custom API base URL |

## Error Handling

Use `errors.As` to distinguish error types:

```go
resp, err := client.UserStatus(ctx, params)
var cfErr *cf.CFError
if errors.As(err, &cfErr) {
    switch cfErr.Code {
    case cf.ErrAPI:
        // Codeforces returned FAILED
    case cf.ErrRateLimit:
        // Rate limited
    case cf.ErrAuth:
        // Auth failed
    }
}
```

## Raw Request

For custom or new API endpoints:

```go
body, err := client.RawRequest(ctx, "user.status", &cf.UserStatusParams{
    Handle: "tourist",
})
```

## API Methods

| Category | Methods |
|----------|---------|
| Blog Entry | `BlogEntryComments`, `BlogEntryView` |
| Contest | `ContestHacks`, `ContestList`, `ContestRatingChanges`, `ContestStandings`, `ContestStatus` |
| ProblemSet | `ProblemsetProblems`, `ProblemsetRecentStatus` |
| User | `UserBlogEntries`, `UserFriends`, `UserInfo`, `UserRatedList`, `UserRating`, `UserStatus`, `UserRecentActions` |
