// Package codeforcesClient 是 Codeforces API 的 Go Client
// 覆盖全部 16 个官方接口，分为 BlogEntry、Contest、ProblemSet、User 四类。
//
// 快速开始：
//
//	client := codeforcesClient.NewClient()
//	resp, _ := client.UserInfo(&codeforcesClient.UserInfoParams{Handles: "tourist"})
//
// 需要认证的接口，设置 Signer 并通过 WithHandle 指定用户：
//
//	client := codeforcesClient.NewClient(
//	    codeforcesClient.WithSigner(codeforcesClient.NewStaticSigner(key, secret)),
//	)
//	resp, _ := client.UserFriends(&codeforcesClient.UserFriendsParams{})
//
// 多用户用 PoolSigner + WithHandle：
//
//	client.WithHandle("alice").UserFriends(&codeforcesClient.UserFriendsParams{})
//	client.WithHandle("bob").UserFriends(&codeforcesClient.UserFriendsParams{})
package codeforcesClient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalhttp "github.com/laoin114514/codeforcesClient/internal/http"
	"github.com/laoin114514/codeforcesClient/internal/params"
)

// apiResponse is the standard Codeforces API response envelope.
// Every API response is wrapped in {"status":"OK"|"FAILED","comment":"...","result":...}.
// 每个 Codeforces API 响应都包裹在 {"status":"OK"|"FAILED","comment":"...","result":...} 信封中。
type apiResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  json.RawMessage `json:"result"`
}

// Client 是 Codeforces API 客户端。
// 使用 NewClient 创建，零值不可用。
type Client struct {
	ctx        context.Context           // 每个 client 的默认 context
	httpClient *http.Client              // 底层 HTTP 客户端
	signer     Signer                    // 可选的请求签名器，用于认证接口
	limiter    *internalhttp.RateLimiter // 限流器，0 = 不限流
	baseURL    string                    // API 基础地址
	transport  *internalhttp.Transport   // 带重试和限流的 HTTP 传输层
}

// NewClient 创建 Client，可传入函数式选项进行配置。
// 默认值：https://codeforces.com/api/、10s 超时、不限流、无认证。
func NewClient(opts ...ClientOption) *Client {
	c := defaultClient()
	for _, opt := range opts {
		opt(c)
	}
	c.transport = internalhttp.NewTransport(c.httpClient, c.limiter)
	return c
}

// WithContext 返回 c 的浅拷贝，使用给定的 context。
// 用于需要超时/取消控制的单次调用。
func (c *Client) WithContext(ctx context.Context) *Client {
	cc := *c
	cc.ctx = ctx
	return &cc
}

// SetContext 直接修改 c 的 context，后续调用都用新 ctx。
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// WithHandle 返回 c 的浅拷贝，在 context 中注入 handle 供 PoolSigner 查找凭据。
//
//	resp, err := client.WithHandle("alice").UserFriends(&cf.UserFriendsParams{})
func (c *Client) WithHandle(handle string) *Client {
	cc := *c
	cc.ctx = WithHandle(c.ctx, handle)
	return &cc
}

// SetHandle 直接修改 c 的 context 注入 handle，后续调用都用此身份。
func (c *Client) SetHandle(handle string) {
	c.ctx = WithHandle(c.ctx, handle)
}

// buildURL 构造完整的 API URL。
// 如果配置了 Signer，会对请求签名并附加认证参数。
func (c *Client) buildURL(ctx context.Context, method string, m map[string]any) (string, error) {
	if c.signer != nil {
		signed, err := c.signer.Sign(ctx, method, m)
		if err != nil {
			return "", &CFError{Code: ErrAuth, Message: "failed to sign request", Cause: err}
		}
		return c.baseURL + signed.URL, nil
	}
	return c.baseURL + method + "?" + params.ToOrderedString(m), nil
}

// doHTTP 发送 GET 请求并返回原始响应体。
// 处理 HTTP 层面的错误：429 限流、4xx/5xx。
func (c *Client) doHTTP(ctx context.Context, urlStr string) ([]byte, error) {
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

	if resp.StatusCode == 429 {
		return nil, &CFError{Code: ErrRateLimit, Message: "rate limited"}
	}
	if resp.StatusCode >= 400 {
		return nil, &CFError{Code: ErrNetwork, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	return body, nil
}

// doRequest 是所有 API 方法的公共请求管线。
// 编码参数 → 构造 URL → 发送请求 → 检查 API 状态 → 解析结果。
func (c *Client) doRequest(ctx context.Context, method string, paramStruct any, extraParams map[string]any, result any) error {
	m, err := params.Encode(paramStruct, extraParams)
	if err != nil {
		return &CFError{Code: ErrInvalidParam, Message: "failed to encode params", Cause: err}
	}

	urlStr, err := c.buildURL(ctx, method, m)
	if err != nil {
		return err
	}

	body, err := c.doHTTP(ctx, urlStr)
	if err != nil {
		return err
	}

	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return &CFError{Code: ErrNetwork, Message: "failed to parse response", Cause: err}
	}
	if envelope.Status != "OK" {
		msg := envelope.Comment
		if msg == "" {
			msg = "API returned status: " + envelope.Status
		}
		return &CFError{Code: ErrAPI, Message: msg}
	}
	if err := json.Unmarshal(body, result); err != nil {
		return &CFError{Code: ErrNetwork, Message: "failed to parse result", Cause: err}
	}
	return nil
}

// RawRequest 发送请求并返回原始 JSON 字节，不做反序列化。
// 适用于未封装的接口或调试场景。
func (c *Client) RawRequest(method string, paramStruct any) ([]byte, error) {
	m, err := params.Encode(paramStruct, nil)
	if err != nil {
		return nil, &CFError{Code: ErrInvalidParam, Message: "failed to encode params", Cause: err}
	}

	urlStr, err := c.buildURL(c.ctx, method, m)
	if err != nil {
		return nil, err
	}

	return c.doHTTP(c.ctx, urlStr)
}

// ==================== Blog Entry ====================

// BlogEntryComments 获取博客文章的评论。
// API: blogEntry.comments
func (c *Client) BlogEntryComments(entryID int) (*BlogEntryCommentsResponse, error) {
	var resp BlogEntryCommentsResponse
	params := &BlogEntryCommentsParams{BlogEntryID: entryID}
	if err := c.doRequest(c.ctx, "blogEntry.comments", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BlogEntryView 获取单篇博客文章。
// API: blogEntry.view
func (c *Client) BlogEntryView(entryID int) (*BlogEntryViewResponse, error) {
	var resp BlogEntryViewResponse
	params := &BlogEntryViewParams{BlogEntryID: entryID}
	if err := c.doRequest(c.ctx, "blogEntry.view", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== Contest ====================

// ContestHacks 获取比赛的 hack 记录。
// API: contest.hacks
func (c *Client) ContestHacks(query *ContestHacksParams) (*ContestHacksResponse, error) {
	var resp ContestHacksResponse
	if err := c.doRequest(c.ctx, "contest.hacks", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ContestList 获取比赛列表。
// API: contest.list
func (c *Client) ContestList(query *ContestListParams) (*ContestListResponse, error) {
	var resp ContestListResponse
	if err := c.doRequest(c.ctx, "contest.list", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ContestRatingChanges 获取比赛后的 Rating 变化。
// API: contest.ratingChanges
func (c *Client) ContestRatingChanges(contestID int) (*ContestRatingChangesResponse, error) {
	var resp ContestRatingChangesResponse
	params := &ContestRatingChangesParams{ContestID: contestID}
	if err := c.doRequest(c.ctx, "contest.ratingChanges", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ContestStandings 获取比赛排名。
// API: contest.standings
func (c *Client) ContestStandings(query *ContestStandingsParams) (*ContestStandingsResponse, error) {
	var resp ContestStandingsResponse
	if err := c.doRequest(c.ctx, "contest.standings", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ContestStatus 获取比赛中的提交记录。
// API: contest.status
func (c *Client) ContestStatus(query *ContestStatusParams) (*ContestStatusResponse, error) {
	var resp ContestStatusResponse
	if err := c.doRequest(c.ctx, "contest.status", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== ProblemSet ====================

// ProblemsetProblems 获取题库中的题目列表。
// API: problemset.problems
func (c *Client) ProblemsetProblems(query *ProblemsetProblemsParams) (*ProblemsetProblemsResponse, error) {
	var resp ProblemsetProblemsResponse
	if err := c.doRequest(c.ctx, "problemset.problems", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ProblemsetRecentStatus 获取题库中最近的提交记录。
// API: problemset.recentStatus
func (c *Client) ProblemsetRecentStatus(query *ProblemsetRecentStatusParams) (*ProblemsetRecentStatusResponse, error) {
	var resp ProblemsetRecentStatusResponse
	if err := c.doRequest(c.ctx, "problemset.recentStatus", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== User ====================

// UserBlogEntries 获取用户的博客文章列表。
// API: user.blogEntries
func (c *Client) UserBlogEntries(query *UserBlogEntriesParams) (*UserBlogEntriesResponse, error) {
	var resp UserBlogEntriesResponse
	if err := c.doRequest(c.ctx, "user.blogEntries", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserFriends 获取当前授权用户的好友列表。
// 需要认证，通过 Signer 提供 API Key。
// API: user.friends
func (c *Client) UserFriends(query *UserFriendsParams) (*UserFriendsResponse, error) {
	var resp UserFriendsResponse
	if err := c.doRequest(c.ctx, "user.friends", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserInfo 根据 handle 获取用户信息。
// API: user.info
func (c *Client) UserInfo(query *UserInfoParams) (*UserInfoResponse, error) {
	var resp UserInfoResponse
	if err := c.doRequest(c.ctx, "user.info", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserRatedList 获取参加过 Rating 的用户列表。
// API: user.ratedList
func (c *Client) UserRatedList(query *UserRatedListParams) (*UserRatedListResponse, error) {
	var resp UserRatedListResponse
	if err := c.doRequest(c.ctx, "user.ratedList", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserRating 获取用户的 Rating 变化历史。
// API: user.rating
func (c *Client) UserRating(query *UserRatingParams) (*UserRatingResponse, error) {
	var resp UserRatingResponse
	if err := c.doRequest(c.ctx, "user.rating", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserStatus 获取用户的提交记录。
// API: user.status
func (c *Client) UserStatus(query *UserStatusParams) (*UserStatusResponse, error) {
	var resp UserStatusResponse
	if err := c.doRequest(c.ctx, "user.status", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RecentActions 获取所有用户最近的动态（博客、评论等）。
// API: recentActions
func (c *Client) RecentActions(maxCount int) (*RecentActionsResponse, error) {
	var resp RecentActionsResponse
	params := &RecentActionsParams{MaxCount: maxCount}
	if err := c.doRequest(c.ctx, "recentActions", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== 其他 ====================

// GroupIsManager 检查指定用户是否为某个组的 manager。
// 需要认证。API: group.isManager
func (c *Client) GroupIsManager(groupCode, handles string) (*GroupIsManagerResponse, error) {
	var resp GroupIsManagerResponse
	params := &GroupIsManagerParams{GroupCode: groupCode, Handles: handles}
	if err := c.doRequest(c.ctx, "group.isManager", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SystemStatus 获取 Codeforces 系统实时健康状态和吞吐量指标。
// API: system.status
func (c *Client) SystemStatus() (*SystemStatusResponse, error) {
	var resp SystemStatusResponse
	if err := c.doRequest(c.ctx, "system.status", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
