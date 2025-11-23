package domain

type BulkDeactivationResult struct {
	DeactivatedUsers []string           `json:"deactivated_users"`
	ReassignedPRs    []ReassignedPR     `json:"reassigned_prs"`
	SkippedPRs       []SkippedPR        `json:"skipped_prs"`
}

type ReassignedPR struct {
	PullRequestID  string `json:"pull_request_id"`
	OldReviewerID  string `json:"old_reviewer_id"`
	NewReviewerID  string `json:"new_reviewer_id"`
}

type SkippedPR struct {
	PullRequestID string `json:"pull_request_id"`
	Reason        string `json:"reason"`
}
