package codeforcesClient

// ==================== Blog Entry ====================
// API: blogEntry.comments, blogEntry.view

// BlogEntryCommentsParams 获取博客评论的参数。
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
// API: contest.hacks, contest.list, contest.ratingChanges, contest.standings, contest.status

// ContestHacksParams 获取比赛 hack 记录的参数。
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
	Status  string           `json:"status"`
	Comment string           `json:"comment,omitempty"`
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
	Status  string        `json:"status"`
	Comment string        `json:"comment,omitempty"`
	Result  []*Submission `json:"result"`
}

// ==================== ProblemSet ====================
// API: problemset.problems, problemset.recentStatus

// ProblemsetProblemsParams 获取题库题目的参数。
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
	Problems          []*Problem           `json:"problems"`
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
// API: user.blogEntries, user.friends, user.info, user.ratedList, user.rating, user.status, user.recentActions

// UserBlogEntriesParams 获取用户博客的参数。
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
	Status  string   `json:"status"`
	Comment string   `json:"comment,omitempty"`
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

// User 表示 Codeforces 用户信息。
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

// RatingChange 表示一次 Rating 变化记录。
type RatingChange struct {
	ContestID               int    `json:"contestId"`
	ContestName             string `json:"contestName"`
	Handle                  string `json:"handle"`
	Rank                    int    `json:"rank"`
	RatingUpdateTimeSeconds int64  `json:"ratingUpdateTimeSeconds"`
	OldRating               int    `json:"oldRating"`
	NewRating               int    `json:"newRating"`
}

// Party 表示参赛方（个人或队伍）。
type Party struct {
	ContestID        int       `json:"contestId,omitempty"`
	Members          []*Member `json:"members"`
	ParticipantType  string    `json:"participantType"`
	TeamID           int       `json:"teamId,omitempty"`
	TeamName         string    `json:"teamName,omitempty"`
	Ghost            bool      `json:"ghost"`
	Room             int       `json:"room,omitempty"`
	StartTimeSeconds int64     `json:"startTimeSeconds,omitempty"`
}

type Member struct {
	Handle string `json:"handle"`
	Name   string `json:"name,omitempty"`
}

// Problem 表示一道题目。
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

// Submission 表示一次提交记录。
type Submission struct {
	ID                  int64     `json:"id"`
	ContestID           int       `json:"contestId,omitempty"`
	CreationTimeSeconds int64     `json:"creationTimeSeconds"`
	RelativeTimeSeconds int64     `json:"relativeTimeSeconds"`
	Problem             *Problem  `json:"problem"`
	Author              *Party    `json:"author"`
	ProgrammingLanguage string    `json:"programmingLanguage"`
	Verdict             string    `json:"verdict,omitempty"`
	Testset             string    `json:"testset"`
	PassedTestCount     int       `json:"passedTestCount"`
	TimeConsumedMillis  int       `json:"timeConsumedMillis"`
	MemoryConsumedBytes int64     `json:"memoryConsumedBytes"`
	Points              float64   `json:"points,omitempty"`
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

// BlogEntry 表示一篇博客文章。
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
	TimeSeconds int64      `json:"timeSeconds"`
	BlogEntry   *BlogEntry `json:"blogEntry,omitempty"`
	Comment     *Comment   `json:"comment,omitempty"`
}

// Contest 表示一场比赛。
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

// RanklistRow 表示排行榜中的一行（一个参赛方）。
type RanklistRow struct {
	Party                      *Party           `json:"party"`
	Rank                       int              `json:"rank"`
	Points                     float64          `json:"points"`
	Penalty                    int              `json:"penalty"`
	SuccessfulHackCount        int              `json:"successfulHackCount"`
	UnsuccessfulHackCount      int              `json:"unsuccessfulHackCount"`
	ProblemResults             []*ProblemResult `json:"problemResults"`
	LastSubmissionTimeSeconds  int64            `json:"lastSubmissionTimeSeconds,omitempty"`
}
