package dto

import "time"

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreateTeamRequest struct {
	TeamName string        `json:"team_name"`
	Members  []*TeamMember `json:"members"`
}

type TeamResponse struct {
	Team *Team `json:"team"`
}

type Team struct {
	TeamName string        `json:"team_name"`
	Members  []*TeamMember `json:"members"`
}

type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User *User `json:"user"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PRResponse struct {
	PR *PullRequest `json:"pr"`
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type ReassignResponse struct {
	PR         *PullRequest `json:"pr"`
	ReplacedBy string       `json:"replaced_by"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type GetReviewResponse struct {
	UserID       string              `json:"user_id"`
	PullRequests []*PullRequestShort `json:"pull_requests"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
