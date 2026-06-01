# Codeforces SDK 重构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有的 URL 拼接工具重构为完整的 Codeforces Go HTTP SDK，覆盖全部 16 个 API 方法。

**Architecture:** 根包暴露 Client + 所有公开类型，`internal/` 下放 HTTP 传输层、签名工具、参数编码三个子包。Client 通过 `doRequest` 统一处理参数编码 → 签名 → HTTP 请求 → 响应解析。

**Tech Stack:** Go 1.24.x, 零外部依赖（仅标准库）

---

### Task 1: 项目初始化与旧代码清理

**Files:**
- Delete: `generator.go`, `hashEncode.go`, `params.go`, `struct_transfer.go`
- Modify: `go.mod`
- Create dirs: `internal/http/`, `internal/signature/`, `internal/params/`

- [ ] **Step 1: 删除旧源文件**

```bash
rm generator.go hashEncode.go params.go struct_transfer.go
```

- [ ] **Step 2: 更新 go.mod**

将 `go.mod` 内容改为：

```
module github.com/laoin114514/codeforcesSDK

go 1.24.0
```

- [ ] **Step 3: 创建 internal 子包目录**

```bash
mkdir -p internal/http internal/signature internal/params
```

- [ ] **Step 4: 验证模块状态**

```bash
go build ./...
```

Expected: 无编译错误

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove old code, set up module and directory structure"
```

---

### Task 2: 错误类型定义 (errors.go)

**Files:**
- Create: `errors.go`

- [ ] **Step 1: 编写 errors.go**

```go
package codeforcessdk

import "fmt"

type ErrorCode int

const (
	ErrNetwork     ErrorCode = iota
	ErrAPI
	ErrRateLimit
	ErrAuth
	ErrInvalidParam
)

type CFError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *CFError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *CFError) Unwrap() error {
	return e.Cause
}

var _ error = (*CFError)(nil)
```

- [ ] **Step 2: 构建验证**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add errors.go
git commit -m "feat: add CFError type with error code classification"
```

---

### Task 3: 参数编码器 (internal/params/)

**Files:**
- Create: `internal/params/encoder.go`
- Create: `internal/params/encoder_test.go`

- [ ] **Step 1: 编写 encoder.go**

```go
package params

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Encode(v any, extra map[string]any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	for k, val := range extra {
		result[k] = val
	}
	return result, nil
}

func ToOrderedString(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprintf("%v", m[k]))
	}
	return b.String()
}
```

- [ ] **Step 2: 编写 encoder_test.go**

```go
package params

import "testing"

type testParams struct {
	Name  string `json:"name"`
	Count int    `json:"count,omitempty"`
	Age   int    `json:"age"`
}

func TestEncode(t *testing.T) {
	p := testParams{Name: "alice", Age: 25}
	m, err := Encode(p, map[string]any{"extra": "val"})
	if err != nil {
		t.Fatal(err)
	}
	if m["name"] != "alice" {
		t.Errorf("name = %v, want alice", m["name"])
	}
	if _, ok := m["count"]; ok {
		t.Error("count should be omitted for zero value")
	}
	if m["age"].(float64) != 25 {
		t.Errorf("age = %v, want 25", m["age"])
	}
	if m["extra"] != "val" {
		t.Errorf("extra = %v, want val", m["extra"])
	}
}

func TestToOrderedString(t *testing.T) {
	got := ToOrderedString(map[string]any{"b": "2", "a": "1", "c": 3})
	want := "a=1&b=2&c=3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToOrderedStringEmpty(t *testing.T) {
	got := ToOrderedString(map[string]any{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./internal/params/ -v
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/params/
git commit -m "feat: add params encoder with ordered string serialization"
```

---

### Task 4: 签名工具 (internal/signature/)

**Files:**
- Create: `internal/signature/hash.go`
- Create: `internal/signature/hash_test.go`

- [ ] **Step 1: 编写 hash.go**

```go
package signature

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func SHA512Sum(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func RandomPrefix() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
```

- [ ] **Step 2: 编写 hash_test.go**

```go
package signature

import "testing"

func TestSHA512Sum(t *testing.T) {
	got := SHA512Sum("hello")
	want := "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRandomPrefix(t *testing.T) {
	for range 10 {
		p := RandomPrefix()
		if len(p) != 6 {
			t.Errorf("len = %d, want 6", len(p))
		}
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./internal/signature/ -v
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/signature/
git commit -m "feat: add SHA512 hash and random prefix utilities"
```

---

### Task 5: Signer 接口与实现 (signer.go)

**Files:**
- Create: `signer.go`

- [ ] **Step 1: 编写 signer.go**

```go
package codeforcessdk

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
```

- [ ] **Step 2: 构建验证**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add signer.go
git commit -m "feat: add Signer interface, StaticSigner and PoolSigner implementations"
```

---

### Task 6: HTTP 传输层 (internal/http/)

**Files:**
- Create: `internal/http/transport.go`
- Create: `internal/http/transport_test.go`

- [ ] **Step 1: 编写 transport.go**

```go
package http

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	rate     int
	interval time.Duration
	mu       sync.Mutex
	last     time.Time
}

func NewRateLimiter(rps int) *RateLimiter {
	r := &RateLimiter{rate: rps}
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
	return r
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.rate <= 0 {
		return nil
	}
	r.mu.Lock()
	elapsed := time.Since(r.last)
	if elapsed < r.interval {
		r.mu.Unlock()
		select {
		case <-time.After(r.interval - elapsed):
		case <-ctx.Done():
			return ctx.Err()
		}
		r.mu.Lock()
	}
	r.last = time.Now()
	r.mu.Unlock()
	return nil
}

func (r *RateLimiter) SetRate(rps int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rate = rps
	if rps > 0 {
		r.interval = time.Second / time.Duration(rps)
	}
}

type Transport struct {
	client     *http.Client
	limiter    *RateLimiter
	maxRetries int
	retryWait  time.Duration
}

func NewTransport(client *http.Client, limiter *RateLimiter) *Transport {
	return &Transport{
		client:     client,
		limiter:    limiter,
		maxRetries: 3,
		retryWait:  500 * time.Millisecond,
	}
}

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

func (t *Transport) isRetryable(code int) bool {
	return code == 429 || code == 503 || code >= 500
}

func (t *Transport) backoff(ctx context.Context, attempt int) {
	wait := t.retryWait * (1 << attempt)
	select {
	case <-time.After(wait):
	case <-ctx.Done():
	}
}
```

- [ ] **Step 2: 编写 transport_test.go**

```go
package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(10)
	start := time.Now()
	for range 3 {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < 200*time.Millisecond {
		t.Errorf("too fast: %v (expected >= 200ms for 3 reqs at 10rps)", elapsed)
	}
}

func TestRateLimiterZeroDisables(t *testing.T) {
	rl := NewRateLimiter(0)
	start := time.Now()
	for range 100 {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("should be instant with rate=0, took %v", elapsed)
	}
}

func TestTransportRetry(t *testing.T) {
	retries := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++
		if retries < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := NewTransport(srv.Client(), NewRateLimiter(0))
	req, _ := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
	resp, err := tr.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if retries != 3 {
		t.Errorf("retries = %d, want 3", retries)
	}
}

func TestRateLimiterContextCancel(t *testing.T) {
	rl := NewRateLimiter(1)
	_ = rl.Wait(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := rl.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./internal/http/ -v -count=1
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/http/
git commit -m "feat: add HTTP transport with rate limiting and retry"
```

---

### Task 7: 所有类型定义 (models.go)

**Files:**
- Create: `models.go`

- [ ] **Step 1: 编写 models.go（Request/Response + 业务对象类型）**

```go
package codeforcessdk

// ==================== Blog Entry ====================

type BlogEntryCommentsParams struct {
	BlogEntryID int `json:"blogEntryId"`
}

type BlogEntryCommentsResponse struct {
	Status  string     `json:"status"`
	Comment string     `json:"comment,omitempty"`
	Result  []*Comment `json:"result"`
}

type BlogEntryViewParams struct {
	BlogEntryID int `json:"blogEntryId"`
}

type BlogEntryViewResponse struct {
	Status  string     `json:"status"`
	Comment string     `json:"comment,omitempty"`
	Result  *BlogEntry `json:"result"`
}

// ==================== Contest ====================

type ContestHacksParams struct {
	ContestID int  `json:"contestId"`
	AsManager bool `json:"asManager,omitempty"`
}

type ContestHacksResponse struct {
	Status  string  `json:"status"`
	Comment string  `json:"comment,omitempty"`
	Result  []*Hack `json:"result"`
}

type ContestListParams struct {
	Gym       bool   `json:"gym,omitempty"`
	GroupCode string `json:"groupCode,omitempty"`
}

type ContestListResponse struct {
	Status  string     `json:"status"`
	Comment string     `json:"comment,omitempty"`
	Result  []*Contest `json:"result"`
}

type ContestRatingChangesParams struct {
	ContestID int `json:"contestId"`
}

type ContestRatingChangesResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  []*RatingChange `json:"result"`
}

type ContestStandingsParams struct {
	ContestID        int    `json:"contestId"`
	AsManager        bool   `json:"asManager,omitempty"`
	From             int    `json:"from,omitempty"`
	Count            int    `json:"count,omitempty"`
	Handles          string `json:"handles,omitempty"`
	Room             int    `json:"room,omitempty"`
	ShowUnofficial   bool   `json:"showUnofficial,omitempty"`
	ParticipantTypes string `json:"participantTypes,omitempty"`
}

type ContestStandingsResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  *StandingsResult `json:"result"`
}

type StandingsResult struct {
	Contest  *Contest       `json:"contest"`
	Problems []*Problem     `json:"problems"`
	Rows     []*RanklistRow `json:"rows"`
}

type ContestStatusParams struct {
	ContestID int    `json:"contestId"`
	Handle    string `json:"handle,omitempty"`
	From      int    `json:"from,omitempty"`
	Count     int    `json:"count,omitempty"`
}

type ContestStatusResponse struct {
	Status  string       `json:"status"`
	Comment string       `json:"comment,omitempty"`
	Result  []*Submission `json:"result"`
}

// ==================== ProblemSet ====================

type ProblemsetProblemsParams struct {
	Tags           string `json:"tags,omitempty"`
	ProblemsetName string `json:"problemsetName,omitempty"`
}

type ProblemsetProblemsResponse struct {
	Status  string            `json:"status"`
	Comment string            `json:"comment,omitempty"`
	Result  *ProblemSetResult `json:"result"`
}

type ProblemSetResult struct {
	Problems          []*Problem          `json:"problems"`
	ProblemStatistics []*ProblemStatistics `json:"problemStatistics"`
}

type ProblemsetRecentStatusParams struct {
	Count          int    `json:"count"`
	ProblemsetName string `json:"problemsetName,omitempty"`
}

type ProblemsetRecentStatusResponse struct {
	Status  string        `json:"status"`
	Comment string        `json:"comment,omitempty"`
	Result  []*Submission `json:"result"`
}

// ==================== User ====================

type UserBlogEntriesParams struct {
	Handle string `json:"handle"`
}

type UserBlogEntriesResponse struct {
	Status  string       `json:"status"`
	Comment string       `json:"comment,omitempty"`
	Result  []*BlogEntry `json:"result"`
}

type UserFriendsParams struct {
	OnlyOnline bool `json:"onlyOnline,omitempty"`
}

type UserFriendsResponse struct {
	Status  string  `json:"status"`
	Comment string  `json:"comment,omitempty"`
	Result  []string `json:"result"`
}

type UserInfoParams struct {
	Handles              string `json:"handles"`
	CheckHistoricHandles bool   `json:"checkHistoricHandles,omitempty"`
}

type UserInfoResponse struct {
	Status  string  `json:"status"`
	Comment string  `json:"comment,omitempty"`
	Result  []*User `json:"result"`
}

type UserRatedListParams struct {
	ActiveOnly           bool `json:"activeOnly,omitempty"`
	IncludeRetired       bool `json:"includeRetired,omitempty"`
	ContestID            int  `json:"contestId,omitempty"`
}

type UserRatedListResponse struct {
	Status  string  `json:"status"`
	Comment string  `json:"comment,omitempty"`
	Result  []*User `json:"result"`
}

type UserRatingParams struct {
	Handle string `json:"handle"`
}

type UserRatingResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  []*RatingChange `json:"result"`
}

type UserStatusParams struct {
	Handle string `json:"handle"`
	From   int    `json:"from,omitempty"`
	Count  int    `json:"count,omitempty"`
}

type UserStatusResponse struct {
	Status  string        `json:"status"`
	Comment string        `json:"comment,omitempty"`
	Result  []*Submission `json:"result"`
}

type RecentActionsParams struct {
	MaxCount int `json:"maxCount"`
}

type RecentActionsResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  []*RecentAction `json:"result"`
}

// ==================== Codeforces 业务对象 ====================

type User struct {
	Handle                  string `json:"handle"`
	Email                   string `json:"email,omitempty"`
	VKID                    string `json:"vkId,omitempty"`
	OpenID                  string `json:"openId,omitempty"`
	FirstName               string `json:"firstName,omitempty"`
	LastName                string `json:"lastName,omitempty"`
	Country                 string `json:"country,omitempty"`
	City                    string `json:"city,omitempty"`
	Organization            string `json:"organization,omitempty"`
	Contribution            int    `json:"contribution"`
	Rank                    string `json:"rank,omitempty"`
	Rating                  int    `json:"rating"`
	MaxRank                 string `json:"maxRank,omitempty"`
	MaxRating               int    `json:"maxRating"`
	LastOnlineTimeSeconds   int64  `json:"lastOnlineTimeSeconds"`
	RegistrationTimeSeconds int64  `json:"registrationTimeSeconds"`
	FriendOfCount           int    `json:"friendOfCount"`
	Avatar                  string `json:"avatar"`
	TitlePhoto              string `json:"titlePhoto"`
}

type RatingChange struct {
	ContestID               int    `json:"contestId"`
	ContestName             string `json:"contestName"`
	Handle                  string `json:"handle"`
	Rank                    int    `json:"rank"`
	RatingUpdateTimeSeconds int64  `json:"ratingUpdateTimeSeconds"`
	OldRating               int    `json:"oldRating"`
	NewRating               int    `json:"newRating"`
}

type Party struct {
	ContestID        int      `json:"contestId,omitempty"`
	Members          []*Member `json:"members"`
	ParticipantType  string   `json:"participantType"`
	TeamID           int      `json:"teamId,omitempty"`
	TeamName         string   `json:"teamName,omitempty"`
	Ghost            bool     `json:"ghost"`
	Room             int      `json:"room,omitempty"`
	StartTimeSeconds int64    `json:"startTimeSeconds,omitempty"`
}

type Member struct {
	Handle string `json:"handle"`
	Name   string `json:"name,omitempty"`
}

type Problem struct {
	ContestID      int      `json:"contestId,omitempty"`
	ProblemsetName string   `json:"problemsetName,omitempty"`
	Index          string   `json:"index"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Points         float64  `json:"points,omitempty"`
	Rating         int      `json:"rating,omitempty"`
	Tags           []string `json:"tags"`
}

type ProblemStatistics struct {
	ContestID   int    `json:"contestId,omitempty"`
	Index       string `json:"index"`
	SolvedCount int    `json:"solvedCount"`
}

type ProblemResult struct {
	Points                    float64 `json:"points"`
	Penalty                   int     `json:"penalty,omitempty"`
	RejectedAttemptCount      int     `json:"rejectedAttemptCount"`
	Type                      string  `json:"type"`
	BestSubmissionTimeSeconds int64   `json:"bestSubmissionTimeSeconds,omitempty"`
}

type Submission struct {
	ID                  int64          `json:"id"`
	ContestID           int            `json:"contestId,omitempty"`
	CreationTimeSeconds int64          `json:"creationTimeSeconds"`
	RelativeTimeSeconds int64          `json:"relativeTimeSeconds"`
	Problem             *Problem       `json:"problem"`
	Author              *Party         `json:"author"`
	ProgrammingLanguage string         `json:"programmingLanguage"`
	Verdict             string         `json:"verdict,omitempty"`
	Testset             string         `json:"testset"`
	PassedTestCount     int            `json:"passedTestCount"`
	TimeConsumedMillis  int            `json:"timeConsumedMillis"`
	MemoryConsumedBytes int64          `json:"memoryConsumedBytes"`
	Points              float64        `json:"points,omitempty"`
}

type Hack struct {
	ID                  int            `json:"id"`
	CreationTimeSeconds int64          `json:"creationTimeSeconds"`
	Hacker              *Party         `json:"hacker"`
	Defender            *Party         `json:"defender"`
	Verdict             string         `json:"verdict,omitempty"`
	Problem             *Problem       `json:"problem"`
	Test                string         `json:"test,omitempty"`
	JudgeProtocol       *JudgeProtocol `json:"judgeProtocol,omitempty"`
}

type JudgeProtocol struct {
	Manual   string `json:"manual"`
	Protocol string `json:"protocol"`
	Verdict  string `json:"verdict"`
}

type BlogEntry struct {
	ID                      int      `json:"id"`
	OriginalLocale          string   `json:"originalLocale"`
	CreationTimeSeconds     int64    `json:"creationTimeSeconds"`
	AuthorHandle            string   `json:"authorHandle"`
	Title                   string   `json:"title"`
	Content                 string   `json:"content,omitempty"`
	Locale                  string   `json:"locale"`
	ModificationTimeSeconds int64    `json:"modificationTimeSeconds"`
	AllowViewHistory        bool     `json:"allowViewHistory"`
	Tags                    []string `json:"tags"`
	Rating                  int      `json:"rating"`
}

type Comment struct {
	ID                  int    `json:"id"`
	CreationTimeSeconds int64  `json:"creationTimeSeconds"`
	CommentatorHandle   string `json:"commentatorHandle"`
	Locale              string `json:"locale"`
	Text                string `json:"text"`
	ParentCommentID     int    `json:"parentCommentId,omitempty"`
	Rating              int    `json:"rating"`
}

type RecentAction struct {
	TimeSeconds int64       `json:"timeSeconds"`
	BlogEntry   *BlogEntry  `json:"blogEntry,omitempty"`
	Comment     *Comment    `json:"comment,omitempty"`
}

type Contest struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Phase               string `json:"phase"`
	Frozen              bool   `json:"frozen"`
	DurationSeconds     int64  `json:"durationSeconds"`
	StartTimeSeconds    int64  `json:"startTimeSeconds,omitempty"`
	RelativeTimeSeconds int64  `json:"relativeTimeSeconds,omitempty"`
	PreparedBy          string `json:"preparedBy,omitempty"`
	WebsiteURL          string `json:"websiteUrl,omitempty"`
	Description         string `json:"description,omitempty"`
	Difficulty          int    `json:"difficulty,omitempty"`
	Kind                string `json:"kind,omitempty"`
	IcpcRegion          string `json:"icpcRegion,omitempty"`
	Country             string `json:"country,omitempty"`
	City                string `json:"city,omitempty"`
	Season              string `json:"season,omitempty"`
}

type RanklistRow struct {
	Party                      *Party          `json:"party"`
	Rank                       int             `json:"rank"`
	Points                     float64         `json:"points"`
	Penalty                    int             `json:"penalty"`
	SuccessfulHackCount        int             `json:"successfulHackCount"`
	UnsuccessfulHackCount      int             `json:"unsuccessfulHackCount"`
	ProblemResults             []*ProblemResult `json:"problemResults"`
	LastSubmissionTimeSeconds  int64           `json:"lastSubmissionTimeSeconds,omitempty"`
}
```

- [ ] **Step 2: 构建验证**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add models.go
git commit -m "feat: add all request/response and business object type definitions"
```

---

### Task 8: Client 选项 (options.go)

**Files:**
- Create: `options.go`

- [ ] **Step 1: 编写 options.go**

```go
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
```

- [ ] **Step 2: 此文件依赖 Client struct（Task 9 中定义），暂不单独构建。Commit**

```bash
git add options.go
git commit -m "feat: add ClientOption functional options"
```

---

### Task 9: Client 实现 (client.go)

**Files:**
- Create: `client.go`

- [ ] **Step 1: 编写 client.go**

```go
package codeforcessdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalhttp "github.com/laoin114514/codeforcesSDK/internal/http"
	"github.com/laoin114514/codeforcesSDK/internal/params"
)

type Client struct {
	httpClient *http.Client
	signer     Signer
	limiter    *internalhttp.RateLimiter
	baseURL    string
	transport  *internalhttp.Transport
}

func NewClient(opts ...ClientOption) *Client {
	c := defaultClient()
	for _, opt := range opts {
		opt(c)
	}
	c.transport = internalhttp.NewTransport(c.httpClient, c.limiter)
	return c
}

// doRequest executes the full request lifecycle: encode → sign → HTTP → parse.
func (c *Client) doRequest(ctx context.Context, method string, paramStruct any, extraParams map[string]any, result any) error {
	m, err := params.Encode(paramStruct, extraParams)
	if err != nil {
		return &CFError{Code: ErrInvalidParam, Message: "failed to encode params", Cause: err}
	}

	var urlStr string
	if c.signer != nil {
		signed, err := c.signer.Sign(ctx, method, m)
		if err != nil {
			return &CFError{Code: ErrAuth, Message: "failed to sign request", Cause: err}
		}
		urlStr = c.baseURL + signed.URL
	} else {
		urlStr = c.baseURL + method + "?" + params.ToOrderedString(m)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return &CFError{Code: ErrNetwork, Message: "failed to create request", Cause: err}
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return &CFError{Code: ErrNetwork, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &CFError{Code: ErrNetwork, Message: "failed to read response body", Cause: err}
	}

	if resp.StatusCode == 429 {
		return &CFError{Code: ErrRateLimit, Message: "rate limited"}
	}
	if resp.StatusCode >= 400 {
		return &CFError{Code: ErrNetwork, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}

	if err := json.Unmarshal(body, result); err != nil {
		return &CFError{Code: ErrNetwork, Message: "failed to parse response", Cause: err}
	}
	return nil
}

func (c *Client) RawRequest(ctx context.Context, method string, paramStruct any) ([]byte, error) {
	m, err := params.Encode(paramStruct, nil)
	if err != nil {
		return nil, &CFError{Code: ErrInvalidParam, Message: "failed to encode params", Cause: err}
	}
	var urlStr string
	if c.signer != nil {
		signed, err := c.signer.Sign(ctx, method, m)
		if err != nil {
			return nil, &CFError{Code: ErrAuth, Message: "failed to sign request", Cause: err}
		}
		urlStr = c.baseURL + signed.URL
	} else {
		urlStr = c.baseURL + method + "?" + params.ToOrderedString(m)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, &CFError{Code: ErrNetwork, Message: "failed to create request", Cause: err}
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, &CFError{Code: ErrNetwork, Message: "request failed", Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &CFError{Code: ErrNetwork, Message: "failed to read response body", Cause: err}
	}
	return body, nil
}

// ==================== Blog Entry ====================

func (c *Client) BlogEntryComments(ctx context.Context, entryID int) (*BlogEntryCommentsResponse, error) {
	var resp BlogEntryCommentsResponse
	params := &BlogEntryCommentsParams{BlogEntryID: entryID}
	if err := c.doRequest(ctx, "blogEntry.comments", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) BlogEntryView(ctx context.Context, entryID int) (*BlogEntryViewResponse, error) {
	var resp BlogEntryViewResponse
	params := &BlogEntryViewParams{BlogEntryID: entryID}
	if err := c.doRequest(ctx, "blogEntry.view", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== Contest ====================

func (c *Client) ContestHacks(ctx context.Context, query *ContestHacksParams) (*ContestHacksResponse, error) {
	var resp ContestHacksResponse
	if err := c.doRequest(ctx, "contest.hacks", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ContestList(ctx context.Context, query *ContestListParams) (*ContestListResponse, error) {
	var resp ContestListResponse
	if err := c.doRequest(ctx, "contest.list", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ContestRatingChanges(ctx context.Context, contestID int) (*ContestRatingChangesResponse, error) {
	var resp ContestRatingChangesResponse
	params := &ContestRatingChangesParams{ContestID: contestID}
	if err := c.doRequest(ctx, "contest.ratingChanges", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ContestStandings(ctx context.Context, query *ContestStandingsParams) (*ContestStandingsResponse, error) {
	var resp ContestStandingsResponse
	if err := c.doRequest(ctx, "contest.standings", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ContestStatus(ctx context.Context, query *ContestStatusParams) (*ContestStatusResponse, error) {
	var resp ContestStatusResponse
	if err := c.doRequest(ctx, "contest.status", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== ProblemSet ====================

func (c *Client) ProblemsetProblems(ctx context.Context, query *ProblemsetProblemsParams) (*ProblemsetProblemsResponse, error) {
	var resp ProblemsetProblemsResponse
	if err := c.doRequest(ctx, "problemset.problems", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ProblemsetRecentStatus(ctx context.Context, query *ProblemsetRecentStatusParams) (*ProblemsetRecentStatusResponse, error) {
	var resp ProblemsetRecentStatusResponse
	if err := c.doRequest(ctx, "problemset.recentStatus", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== User ====================

func (c *Client) UserBlogEntries(ctx context.Context, query *UserBlogEntriesParams) (*UserBlogEntriesResponse, error) {
	var resp UserBlogEntriesResponse
	if err := c.doRequest(ctx, "user.blogEntries", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserFriends(ctx context.Context, query *UserFriendsParams) (*UserFriendsResponse, error) {
	var resp UserFriendsResponse
	if err := c.doRequest(ctx, "user.friends", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserInfo(ctx context.Context, query *UserInfoParams) (*UserInfoResponse, error) {
	var resp UserInfoResponse
	if err := c.doRequest(ctx, "user.info", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserRatedList(ctx context.Context, query *UserRatedListParams) (*UserRatedListResponse, error) {
	var resp UserRatedListResponse
	if err := c.doRequest(ctx, "user.ratedList", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserRating(ctx context.Context, query *UserRatingParams) (*UserRatingResponse, error) {
	var resp UserRatingResponse
	if err := c.doRequest(ctx, "user.rating", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserStatus(ctx context.Context, query *UserStatusParams) (*UserStatusResponse, error) {
	var resp UserStatusResponse
	if err := c.doRequest(ctx, "user.status", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UserRecentActions(ctx context.Context, maxCount int) (*RecentActionsResponse, error) {
	var resp RecentActionsResponse
	params := &RecentActionsParams{MaxCount: maxCount}
	if err := c.doRequest(ctx, "user.recentActions", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 2: 构建验证**

```bash
go build ./...
```

Expected: success

- [ ] **Step 3: Commit**

```bash
git add client.go
git commit -m "feat: add Client with all 16 Codeforces API methods and RawRequest"
```

---

### Task 10: 集成测试

**Files:**
- Create: `codeforces_test.go`

- [ ] **Step 1: 编写集成测试**

```go
//go:build integration

package codeforcessdk

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationUserInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.UserInfo(ctx, &UserInfoParams{Handles: "tourist"})
	if err != nil {
		t.Fatalf("UserInfo failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK (comment: %s)", resp.Status, resp.Comment)
	}
	if len(resp.Result) == 0 {
		t.Fatal("no users returned")
	}
	if resp.Result[0].Handle != "tourist" {
		t.Errorf("handle = %s, want tourist", resp.Result[0].Handle)
	}
}

func TestIntegrationUserRating(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.UserRating(ctx, &UserRatingParams{Handle: "tourist"})
	if err != nil {
		t.Fatalf("UserRating failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
}

func TestIntegrationContestList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.ContestList(ctx, &ContestListParams{Gym: false})
	if err != nil {
		t.Fatalf("ContestList failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
	if len(resp.Result) == 0 {
		t.Fatal("no contests returned")
	}
}

func TestIntegrationWithAuth(t *testing.T) {
	apiKey := os.Getenv("CF_API_KEY")
	secret := os.Getenv("CF_SECRET")
	if apiKey == "" || secret == "" {
		t.Skip("CF_API_KEY and CF_SECRET not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient(WithSigner(NewStaticSigner(apiKey, secret)))
	resp, err := client.UserFriends(ctx, &UserFriendsParams{})
	if err != nil {
		t.Fatalf("UserFriends failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
}
```

- [ ] **Step 2: 验证普通测试不包含集成测试**

```bash
go test ./... -v -count=1
```

Expected: 运行 internal 子包的单元测试，跳过集成测试

- [ ] **Step 3: 运行集成测试（可选）**

```bash
go test -tags=integration -v -count=1 ./...
```

- [ ] **Step 4: Commit**

```bash
git add codeforces_test.go
git commit -m "test: add integration tests for major API methods"
```

---

### Task 11: README.md

**Files:**
- Create: `README.md`

- [ ] **Step 1: 编写 README**

````markdown
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

## API Methods

| Category | Methods |
|----------|---------|
| Blog Entry | `BlogEntryComments`, `BlogEntryView` |
| Contest | `ContestHacks`, `ContestList`, `ContestRatingChanges`, `ContestStandings`, `ContestStatus` |
| ProblemSet | `ProblemsetProblems`, `ProblemsetRecentStatus` |
| User | `UserBlogEntries`, `UserFriends`, `UserInfo`, `UserRatedList`, `UserRating`, `UserStatus`, `UserRecentActions` |

Plus `RawRequest(ctx, method, params)` for custom API calls.
````

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with quick start and API reference"
```

---

### Task 12: 最终构建验证

- [ ] **Step 1: 运行所有单元测试**

```bash
go test ./... -v -count=1
```

Expected: all PASS

- [ ] **Step 2: 验证 go vet**

```bash
go vet ./...
```

Expected: no warnings

- [ ] **Step 3: 最终提交（如有遗漏文件）**

```bash
git add -A
git status
```
