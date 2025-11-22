package handlers

import (
	"context"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
)

type TeamService interface {
	CreateTeam(ctx context.Context, teamName string, members []*domain.User) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserService interface {
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

type PRService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error)
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
}
