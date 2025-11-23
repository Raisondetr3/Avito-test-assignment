package domain

type Statistics struct {
	PullRequests PullRequestStats `json:"pull_requests"`
	Users        UserStats        `json:"users"`
	Teams        TeamStats        `json:"teams"`
	TopReviewers []ReviewerStat   `json:"top_reviewers"`
}

type PullRequestStats struct {
	Total  int `json:"total"`
	Open   int `json:"open"`
	Merged int `json:"merged"`
}

type UserStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
}

type TeamStats struct {
	Total int `json:"total"`
}

type ReviewerStat struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	ReviewCount int    `json:"review_count"`
}
