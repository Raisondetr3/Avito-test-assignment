package domain

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

func (s PRStatus) String() string {
	return string(s)
}

func (s PRStatus) IsValid() bool {
	return s == PRStatusOpen || s == PRStatusMerged
}

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MergedAt          *time.Time
}

func NewPullRequest(pullRequestID, pullRequestName, authorID string) *PullRequest {
	now := time.Now()
	return &PullRequest{
		PullRequestID:     pullRequestID,
		PullRequestName:   pullRequestName,
		AuthorID:          authorID,
		Status:            PRStatusOpen,
		AssignedReviewers: make([]string, 0, 2),
		CreatedAt:         now,
		UpdatedAt:         now,
		MergedAt:          nil,
	}
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

func (pr *PullRequest) IsOpen() bool {
	return pr.Status == PRStatusOpen
}

func (pr *PullRequest) CanModifyReviewers() bool {
	return pr.IsOpen()
}

func (pr *PullRequest) Merge() {
	if pr.IsMerged() {
		return
	}

	now := time.Now()
	pr.Status = PRStatusMerged
	pr.MergedAt = &now
	pr.UpdatedAt = now
}

func (pr *PullRequest) AssignReviewers(reviewerIDs []string) {
	if len(reviewerIDs) > 2 {
		reviewerIDs = reviewerIDs[:2]
	}
	pr.AssignedReviewers = reviewerIDs
	pr.UpdatedAt = time.Now()
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string) bool {
	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			pr.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

func (pr *PullRequest) GetReviewerCount() int {
	return len(pr.AssignedReviewers)
}
