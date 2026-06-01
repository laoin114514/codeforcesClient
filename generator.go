package codeforcessdk

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ============================CF URL生成器===============================================
const BaseUrl = "https://codeforces.com/api/"

type GenerateCFurl struct {
	mu         sync.RWMutex
	baseUrl    string
	User       *user
	Contest    *contest
	ProblemSet *problemSet
	apiKeyPool map[string]*UserApikey
}

type UserApikey struct {
	Apikey    string
	SecretKey string
}

type user struct {
	g *GenerateCFurl
}
type contest struct {
	g *GenerateCFurl
}
type problemSet struct {
	g *GenerateCFurl
}

func NewGenerator(apiKeyPool map[string]*UserApikey) *GenerateCFurl {
	g := &GenerateCFurl{
		baseUrl:    BaseUrl,
		User:       &user{},
		ProblemSet: &problemSet{},
		Contest:    &contest{},
		apiKeyPool: apiKeyPool,
	}
	g.User.g = g
	g.Contest.g = g
	g.ProblemSet.g = g
	return g
}
func (g *GenerateCFurl) SetApiKeyPool(apiKeyPool map[string]*UserApikey) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.apiKeyPool = apiKeyPool
}
func (g *GenerateCFurl) getApikeyFromPool(handle string) (string, string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if handle == "" {
		return "", "", errors.New("handle不能为空")
	}
	userApiKey, ok := g.apiKeyPool[handle]
	if ok && userApiKey.Apikey != "" && userApiKey.SecretKey != "" {
		return userApiKey.Apikey, userApiKey.SecretKey, nil
	}
	return "", "", errors.New("apikey不存在")
}

// 组合url和apikey
func (g *GenerateCFurl) combineUrlWithApikey(handle string, method string, pararms any) (string, error) {
	baseUrl := BaseUrl
	apikey, secret, err := g.getApikeyFromPool(handle)
	if err != nil {
		return "", err
	}
	now := time.Now()
	time := now.Unix()
	randomKey := g.randomNumber(6)
	//将参数转换为字符串
	pararmStr, err := NewStructTransfer(pararms).
		AddParam("apiKey", apikey).
		AddParam("time", time).
		ToOrderStr()
	if err != nil {
		return "", err
	}
	tail := fmt.Sprintf("%v?%v", method, pararmStr)
	hashCode := NewHashEncoder().Hash512(fmt.Sprintf("%v/%v#%v", randomKey, tail, secret))
	url := fmt.Sprintf("%v%v&apiSig=%v%v", baseUrl, tail, randomKey, hashCode)
	return url, nil
}
func (g *GenerateCFurl) combineUrlWithNoApikey(method string, pararms any) (string, error) {
	baseUrl := BaseUrl
	pararmStr, err := NewStructTransfer(pararms).
		ToOrderStr()
	if err != nil {
		return "", err
	}
	tail := fmt.Sprintf("%v?%v", method, pararmStr)
	return fmt.Sprintf("%v%v", baseUrl, tail), nil
}
func (g *GenerateCFurl) randomNumber(n int) string {
	return fmt.Sprintf("%06d", rand.Intn(int(math.Pow10(n))))
}

// ============================================User============================================//
func (u *user) Status(useApikey bool, query *UserStatusParams) (string, error) {
	if useApikey {
		return u.g.combineUrlWithApikey(query.Handle, "user.status", query)
	}
	return u.g.combineUrlWithNoApikey("user.status", query)
}
func (u *user) Rating(query *UserRatingParams) (string, error) {
	return u.g.combineUrlWithApikey(query.Handle, "user.rating", query)
}
func (u *user) RatedList(query *UserRatedListParams) (string, error) {
	return u.g.combineUrlWithNoApikey("user.ratedList", query)
}

func (u *user) Info(query *UserInfoParams) (string, error) {
	return u.g.combineUrlWithNoApikey("user.info", query)
}

func (u *user) Friends(handle string, query *UserFriendsParams) (string, error) {
	return u.g.combineUrlWithApikey(handle, "user.friends", query)
}

func (u *user) BlogEntries(query *UserBlogEntriesParams) (string, error) {
	return u.g.combineUrlWithApikey(query.Handle, "user.blogEntries", query)
}

func (u *user) RecentActions(query *RecentActionsParams) (string, error) {
	return u.g.combineUrlWithNoApikey("user.recentActions", query)
}

// ============================================Contest============================================//
func (c *contest) List(handle string, query *ContestListParams) (string, error) {
	return c.g.combineUrlWithApikey(handle, "contest.list", query)
}
func (c *contest) Standings(handle string, query *ContestStandingsParams) (string, error) {
	return c.g.combineUrlWithApikey(handle, "contest.standings", query)
}
func (c *contest) Status(handle string, query *ContestStatusParams) (string, error) {
	return c.g.combineUrlWithApikey(handle, "contest.status", query)
}

// ============================================ProblemSet============================================//
func (p *problemSet) Problems(query *ProblemsetProblemsParams) (string, error) {
	return p.g.combineUrlWithNoApikey("problemset.problems", query)
}
func (p *problemSet) RecentStatus(query *ProblemsetRecentStatusParams) (string, error) {
	return p.g.combineUrlWithNoApikey("problemset.recentStatus", query)
}
