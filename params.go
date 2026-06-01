package codeforcessdk

// ==================== Codeforces API 请求参数结构体 ====================

// BlogEntryCommentsParams 博客评论请求参数
type BlogEntryCommentsParams struct {
	BlogEntryID int `json:"blogEntryId"` // 博客条目ID，必填
}

// BlogEntryViewParams 博客查看请求参数
type BlogEntryViewParams struct {
	BlogEntryID int `json:"blogEntryId"` // 博客条目ID，必填
}

// ContestHacksParams 比赛hack请求参数
type ContestHacksParams struct {
	ContestID int  `json:"contestId"`           // 比赛ID，必填
	AsManager bool `json:"asManager,omitempty"` // 是否以管理员身份查看，可选
}

// ContestListParams 比赛列表请求参数
type ContestListParams struct {
	Gym       bool   `json:"gym,omitempty"`       // 是否返回Gym比赛，可选
	GroupCode string `json:"groupCode,omitempty"` // 用于筛选比赛的组代码，可选
}

// ContestRatingChangesParams 比赛评分变化请求参数
type ContestRatingChangesParams struct {
	ContestID int `json:"contestId"` // 比赛ID，必填
}

// ContestStandingsParams 比赛排名请求参数
type ContestStandingsParams struct {
	ContestID        int    `json:"contestId"`                  // 比赛ID，必填
	AsManager        bool   `json:"asManager,omitempty"`        // 是否以管理员身份查看，可选
	From             int    `json:"from,omitempty"`             // 排名列表的起始索引（1-based），可选
	Count            int    `json:"count,omitempty"`            // 返回的排名行数，可选
	Handles          string `json:"handles,omitempty"`          // 分号分隔的用户句柄列表，可选
	Room             int    `json:"room,omitempty"`             // 房间号，可选
	ShowUnofficial   bool   `json:"showUnofficial,omitempty"`   // 是否显示非正式参赛者，可选
	ParticipantTypes string `json:"participantTypes,omitempty"` // 参赛者类型的逗号分隔列表，可选
}

// ContestStatusParams 比赛状态请求参数
type ContestStatusParams struct {
	ContestID int    `json:"contestId"`        // 比赛ID，必填
	Handle    string `json:"handle,omitempty"` // 用户句柄，可选
	From      int    `json:"from,omitempty"`   // 提交的起始索引（1-based），可选
	Count     int    `json:"count,omitempty"`  // 返回的提交数，可选
}

// ProblemsetProblemsParams 题目集题目请求参数
type ProblemsetProblemsParams struct {
	Tags           string `json:"tags,omitempty"`           // 分号分隔的标签列表，可选
	ProblemsetName string `json:"problemsetName,omitempty"` // 自定义题集的短名称，可选
}

// ProblemsetRecentStatusParams 题目集最近状态请求参数
type ProblemsetRecentStatusParams struct {
	Count          int    `json:"count"`                    // 返回的提交数，必填
	ProblemsetName string `json:"problemsetName,omitempty"` // 自定义题集的短名称，可选
}

// RecentActionsParams 最近操作请求参数
type RecentActionsParams struct {
	MaxCount int `json:"maxCount"` // 返回的最近操作数，必填
}

// UserBlogEntriesParams 用户博客条目请求参数
type UserBlogEntriesParams struct {
	Handle string `json:"handle"` // 用户句柄，必填
}

// UserFriendsParams 用户好友请求参数
type UserFriendsParams struct {
	OnlyOnline bool `json:"onlyOnline,omitempty"` // 是否仅返回在线好友，可选
}

// UserInfoParams 用户信息请求参数
type UserInfoParams struct {
	Handles              string `json:"handles"`                        // 分号分隔的用户句柄列表，必填
	CheckHistoricHandles bool   `json:"checkHistoricHandles,omitempty"` // 是否检查历史句柄，可选
}

// UserRatedListParams 用户评分列表请求参数
type UserRatedListParams struct {
	ActiveOnly bool `json:"activeOnly,omitempty"` // 是否仅返回活跃用户，可选
}

// UserRatingParams 用户评分请求参数
type UserRatingParams struct {
	Handle string `json:"handle"` // 用户句柄，必填
}

// UserStatusParams 用户状态请求参数
type UserStatusParams struct {
	Handle string `json:"handle"`          // 用户句柄，必填
	From   int    `json:"from,omitempty"`  // 提交的起始索引（1-based），可选
	Count  int    `json:"count,omitempty"` // 返回的提交数，可选
}
