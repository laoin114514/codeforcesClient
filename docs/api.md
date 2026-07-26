# Codeforces Go SDK API 文档

Codeforces API 的 Go SDK，覆盖全部 18 个官方接口。

## 安装

```bash
go get github.com/laoin114514/codeforcesClient
```

## 导入

```go
import cf "github.com/laoin114514/codeforcesClient"
```

---

## 创建客户端

```go
// 默认客户端：codeforces.com/api、10s 超时、不限流、无认证
client := cf.NewClient()

// 自定义配置
client := cf.NewClient(
    cf.WithHTTPClient(customHTTPClient),   // 自定义 HTTP 客户端
    cf.WithSigner(signer),                 // 认证签名器
    cf.WithRateLimit(5),                   // 每秒最多 5 次请求
    cf.WithBaseURL("https://custom.api/"), // 自定义 API 地址
    cf.WithContext(ctx),                   // 默认 context
)
```

### 选项说明

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `WithHTTPClient` | `*http.Client` | `&http.Client{Timeout: 10s}` | 自定义 HTTP 客户端 |
| `WithSigner` | `Signer` | `nil` | 认证签名器，用于需要登录的接口 |
| `WithRateLimit` | `int` | `0`（不限） | 每秒最大请求数 |
| `WithBaseURL` | `string` | `https://codeforces.com/api/` | API 基础地址 |
| `WithContext` | `context.Context` | `context.Background()` | 默认 context |

---

## Context 和超时

Client 内置默认 context（`context.Background()`），普通调用无需传 ctx：

```go
resp, _ := client.UserInfo(&cf.UserInfoParams{Handles: "tourist"})
```

需要超时或取消时，用 `WithContext` 派生一个临时 client：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := client.WithContext(ctx).UserInfo(&cf.UserInfoParams{Handles: "tourist"})
```

---

## 认证

需要授权的接口（如 `UserFriends`、`GroupIsManager`）通过 `Signer` 传入 API Key。

### 单用户：StaticSigner

```go
client := cf.NewClient(cf.WithSigner(cf.NewStaticSigner("your-api-key", "your-secret")))
resp, _ := client.UserFriends(&cf.UserFriendsParams{})
```

### 多用户：PoolSigner

通过 context 传递 handle 来切换用户凭据：

```go
signer := cf.NewPoolSigner(map[string]struct{ ApiKey, Secret string }{
    "user1": {ApiKey: "key1", Secret: "secret1"},
    "user2": {ApiKey: "key2", Secret: "secret2"},
})
client := cf.NewClient(cf.WithSigner(signer))

// 以 user1 身份调用
ctx := cf.WithHandle(context.Background(), "user1")
resp, _ := client.WithContext(ctx).UserFriends(&cf.UserFriendsParams{})
```

---

## 错误处理

所有方法都返回 `*CFError`，用 `errors.As` 提取并区分类型：

```go
resp, err := client.UserStatus(&cf.UserStatusParams{Handle: "tourist"})
if err != nil {
    var cfErr *cf.CFError
    if errors.As(err, &cfErr) {
        switch cfErr.Code {
        case cf.ErrAPI:
            fmt.Println("Codeforces 返回 FAILED:", cfErr.Message)
        case cf.ErrRateLimit:
            fmt.Println("触发限流，稍后重试")
        case cf.ErrAuth:
            fmt.Println("认证失败:", cfErr.Message)
        case cf.ErrInvalidParam:
            fmt.Println("参数无效:", cfErr.Message)
        case cf.ErrNetwork:
            fmt.Println("网络错误:", cfErr.Message)
        }
    }
    return
}
```

---

## API 方法

### Blog Entry（博客）

#### BlogEntryComments

获取博客文章的评论列表。

```go
resp, err := client.BlogEntryComments(1)
// resp.Result: []*Comment — 评论列表
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `entryID` | `int` | 博客文章 ID |

#### BlogEntryView

获取单篇博客文章的完整内容。

```go
resp, err := client.BlogEntryView(1)
// resp.Result: *BlogEntry — 文章详情（含标题、正文、标签等）
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `entryID` | `int` | 博客文章 ID |

---

### Contest（比赛）

#### ContestHacks

获取指定比赛的 hack 记录。比赛结束后返回完整数据，进行中仅返回自己的 hack。

```go
resp, err := client.ContestHacks(&cf.ContestHacksParams{
    ContestID: 1,
})
// resp.Result: []*Hack — hack 记录列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ContestID` | `int` | 是 | 比赛 ID |
| `AsManager` | `bool` | 否 | 是否以 manager 身份查看 |

#### ContestList

获取比赛列表。

```go
resp, err := client.ContestList(&cf.ContestListParams{Gym: false})
// resp.Result: []*Contest — 比赛列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Gym` | `bool` | 否 | 是否只返回 gym 比赛 |
| `GroupCode` | `string` | 否 | 组代码，筛选特定组的比赛 |

#### ContestRatingChanges

获取比赛后的 Rating 变化。

```go
resp, err := client.ContestRatingChanges(1)
// resp.Result: []*RatingChange — Rating 变化列表
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `contestID` | `int` | 比赛 ID |

#### ContestStandings

获取比赛排名和题目列表。

```go
resp, err := client.ContestStandings(&cf.ContestStandingsParams{
    ContestID:      1,
    From:           0,
    Count:          10,
    Handles:        "tourist;Petr",
    ShowUnofficial: true,
})
// resp.Result.Contest  — 比赛信息
// resp.Result.Problems — 题目列表
// resp.Result.Rows     — 排名行列表
```

> **注意**：非 gym 比赛只能用匿名请求访问，且不能携带额外参数。带认证的请求仅支持 gym 比赛。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ContestID` | `int` | 是 | 比赛 ID |
| `From` | `int` | 否 | 起始排名（1-indexed） |
| `Count` | `int` | 否 | 返回行数 |
| `Handles` | `string` | 否 | 筛选指定用户，分号分隔 |
| `Room` | `int` | 否 | 筛选指定房间 |
| `ShowUnofficial` | `bool` | 否 | 是否显示非正式参赛者 |
| `ParticipantTypes` | `string` | 否 | 筛选参赛类型，分号分隔 |
| `AsManager` | `bool` | 否 | 以 manager 身份查看 |

#### ContestStatus

获取比赛中的提交记录。

```go
resp, err := client.ContestStatus(&cf.ContestStatusParams{
    ContestID: 1,
    Handle:    "tourist",
    From:      0,
    Count:     10,
})
// resp.Result: []*Submission — 提交列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ContestID` | `int` | 是 | 比赛 ID |
| `Handle` | `string` | 否 | 筛选指定用户 |
| `From` | `int` | 否 | 起始位置（1-indexed） |
| `Count` | `int` | 否 | 返回数量 |

---

### ProblemSet（题库）

#### ProblemsetProblems

获取题库中的题目列表，可按标签筛选。

```go
resp, err := client.ProblemsetProblems(&cf.ProblemsetProblemsParams{
    Tags:           "implementation;dp",
    ProblemsetName: "acmsguru",
})
// resp.Result.Problems          — 题目列表
// resp.Result.ProblemStatistics — 各题解题统计
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Tags` | `string` | 否 | 标签筛选，分号分隔 |
| `ProblemsetName` | `string` | 否 | 题库名称 |

#### ProblemsetRecentStatus

获取题库中最近的提交记录。

```go
resp, err := client.ProblemsetRecentStatus(&cf.ProblemsetRecentStatusParams{
    Count:          10,
    ProblemsetName: "acmsguru",
})
// resp.Result: []*Submission — 最近的提交列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Count` | `int` | 是 | 返回数量（最大 1000） |
| `ProblemsetName` | `string` | 否 | 题库名称 |

---

### User（用户）

#### UserInfo

根据 handle 获取用户信息。

```go
resp, err := client.UserInfo(&cf.UserInfoParams{
    Handles:              "tourist;Petr;jiangly",
    CheckHistoricHandles: true,
})
// resp.Result: []*User — 用户信息列表
for _, u := range resp.Result {
    fmt.Printf("%s: rating=%d, rank=%s\n", u.Handle, u.Rating, u.Rank)
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Handles` | `string` | 是 | Codeforces handle，分号分隔 |
| `CheckHistoricHandles` | `bool` | 否 | 是否查找历史 handle |

#### UserRating

获取用户的 Rating 变化历史。

```go
resp, err := client.UserRating(&cf.UserRatingParams{Handle: "tourist"})
// resp.Result: []*RatingChange — Rating 变化列表
for _, r := range resp.Result {
    fmt.Printf("%s: %d → %d (rank %d)\n",
        r.ContestName, r.OldRating, r.NewRating, r.Rank)
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Handle` | `string` | 是 | Codeforces handle |

#### UserStatus

获取用户的提交记录。

```go
resp, err := client.UserStatus(&cf.UserStatusParams{
    Handle: "tourist",
    From:   0,
    Count:  10,
})
// resp.Result: []*Submission — 提交列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Handle` | `string` | 是 | Codeforces handle |
| `From` | `int` | 否 | 起始位置（1-indexed） |
| `Count` | `int` | 否 | 返回数量 |

#### UserRatedList

获取参加过 Rating 的用户列表。

```go
resp, err := client.UserRatedList(&cf.UserRatedListParams{
    ActiveOnly:     true,
    IncludeRetired: false,
})
// resp.Result: []*User — 用户列表（按 rating 降序）
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ActiveOnly` | `bool` | 否 | 仅在过去一个月活跃的用户 |
| `IncludeRetired` | `bool` | 否 | 包含退役用户 |
| `ContestID` | `int` | 否 | 按特定比赛筛选 |

#### UserBlogEntries

获取用户的博客文章列表。

```go
resp, err := client.UserBlogEntries(&cf.UserBlogEntriesParams{Handle: "tourist"})
// resp.Result: []*BlogEntry — 博客列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Handle` | `string` | 是 | Codeforces handle |

#### UserFriends

获取当前授权用户的好友列表。

> **需要认证**

```go
client := cf.NewClient(cf.WithSigner(cf.NewStaticSigner(key, secret)))
resp, err := client.UserFriends(&cf.UserFriendsParams{OnlyOnline: true})
// resp.Result: []string — 好友 handle 列表
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `OnlyOnline` | `bool` | 否 | 仅返回在线好友 |

---

### 其他

#### RecentActions

获取全站最近的动态（博客发布、评论等）。

```go
resp, err := client.RecentActions(10)
// resp.Result: []*RecentAction — 动态列表
// 每条包含 TimeSeconds + BlogEntry/Comment 之一
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `maxCount` | `int` | 返回数量（最大 100） |

#### GroupIsManager

检查指定用户是否为某个组的 manager。

> **需要认证**

```go
resp, err := client.GroupIsManager("GROUP_CODE", "user1;user2")
// resp.Result: []string — 是 manager 的 handle 列表
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `groupCode` | `string` | 组代码 |
| `handles` | `string` | 要检查的 handle，分号分隔 |

#### SystemStatus

获取 Codeforces 系统实时健康状态和吞吐量指标。

```go
resp, err := client.SystemStatus()
// resp.Result.LangVersion — 语言环境版本
// resp.Result.RPS         — 当前每秒处理请求数
// resp.Result.Now         — 服务器时间戳
```

无需参数。

---

### RawRequest

对于未封装的接口或需要原始响应的场景：

```go
body, err := client.RawRequest("user.info", &cf.UserInfoParams{Handles: "tourist"})
// body: []byte — Codeforces API 的原始 JSON 响应
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `method` | `string` | API 方法名，如 `"user.info"` |
| `paramStruct` | `any` | 参数结构体或 `nil` |

---

## 业务对象速查

### User — 用户信息

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| `Handle` | `string` | `handle` | Codeforces 用户名 |
| `Rating` | `int` | `rating` | 当前 Rating |
| `MaxRating` | `int` | `maxRating` | 历史最高 Rating |
| `Rank` | `string` | `rank` | 头衔（如 "legendary grandmaster"） |
| `MaxRank` | `string` | `maxRank` | 历史最高头衔 |
| `Contribution` | `int` | `contribution` | 社区贡献值 |
| `FirstName` | `string` | `firstName` | 名 |
| `LastName` | `string` | `lastName` | 姓 |
| `Country` | `string` | `country` | 国家 |
| `Organization` | `string` | `organization` | 学校/组织 |
| `Avatar` | `string` | `avatar` | 头像 URL |
| `FriendOfCount` | `int` | `friendOfCount` | 被多少人加为好友 |

### Submission — 提交记录

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `int64` | 提交 ID |
| `Problem` | `*Problem` | 题目 |
| `Author` | `*Party` | 提交者 |
| `Verdict` | `string` | 评测结果（如 "OK", "WRONG_ANSWER"） |
| `ProgrammingLanguage` | `string` | 编程语言 |
| `PassedTestCount` | `int` | 通过的测试数 |
| `TimeConsumedMillis` | `int` | 耗时（毫秒） |
| `MemoryConsumedBytes` | `int64` | 内存消耗（字节） |
| `CreationTimeSeconds` | `int64` | 提交时间戳 |

### RatingChange — Rating 变化

| 字段 | 类型 | 说明 |
|------|------|------|
| `ContestID` | `int` | 比赛 ID |
| `ContestName` | `string` | 比赛名称 |
| `OldRating` | `int` | 变化前 Rating |
| `NewRating` | `int` | 变化后 Rating |
| `Rank` | `int` | 比赛排名 |

### Contest — 比赛

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `int` | 比赛 ID |
| `Name` | `string` | 比赛名称 |
| `Phase` | `string` | 阶段：`"BEFORE"`, `"CODING"`, `"FINISHED"` 等 |
| `DurationSeconds` | `int64` | 时长（秒） |
| `StartTimeSeconds` | `int64` | 开始时间戳 |

### Problem — 题目

| 字段 | 类型 | 说明 |
|------|------|------|
| `ContestID` | `int` | 所属比赛 ID |
| `Index` | `string` | 题目编号（如 "A", "B2"） |
| `Name` | `string` | 题目标题 |
| `Rating` | `int` | 难度等级 |
| `Tags` | `[]string` | 题目标签列表 |
