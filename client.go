package codeforcesClient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalhttp "github.com/laoin114514/codeforcesSDK/internal/http"
	"github.com/laoin114514/codeforcesSDK/internal/params"
)

// apiResponse is the standard Codeforces API envelope.
type apiResponse struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  json.RawMessage `json:"result"`
}

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

func (c *Client) RawRequest(ctx context.Context, method string, paramStruct any) ([]byte, error) {
	m, err := params.Encode(paramStruct, nil)
	if err != nil {
		return nil, &CFError{Code: ErrInvalidParam, Message: "failed to encode params", Cause: err}
	}

	urlStr, err := c.buildURL(ctx, method, m)
	if err != nil {
		return nil, err
	}

	return c.doHTTP(ctx, urlStr)
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
