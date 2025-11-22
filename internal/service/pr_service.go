package service

import (
	"context"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
)

type PRRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	Update(ctx context.Context, pr *domain.PullRequest) error
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
}

type PRService struct {
	prRepo       PRRepository
	userRepo     UserRepository
	reviewerAssg *ReviewerAssigner
}

func NewPRService(prRepo PRRepository, userRepo UserRepository, reviewerAssg *ReviewerAssigner) *PRService {
	return &PRService{
		prRepo:       prRepo,
		userRepo:     userRepo,
		reviewerAssg: reviewerAssg,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.ErrPRExists(prID)
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	candidates, err := s.userRepo.GetActiveByTeamExcluding(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	pr := domain.NewPullRequest(prID, prName, authorID)

	reviewerIDs := s.reviewerAssg.SelectReviewers(candidates, 2)
	pr.AssignReviewers(reviewerIDs)

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	pr.Merge()

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.IsMerged() {
		return nil, "", errors.ErrPRMerged(prID)
	}

	if !pr.HasReviewer(oldReviewerID) {
		return nil, "", errors.ErrNotAssigned(oldReviewerID, prID)
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	candidates, err := s.userRepo.GetActiveByTeamExcluding(ctx, oldReviewer.TeamName, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	newReviewerID, found := s.reviewerAssg.SelectRandomReviewer(candidates)
	if !found {
		return nil, "", errors.ErrNoCandidate(oldReviewer.TeamName)
	}

	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return nil, "", err
	}

	pr, err = s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

func (s *PRService) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error) {
	return s.prRepo.GetByReviewer(ctx, reviewerID)
}

func (s *PRService) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.prRepo.GetByID(ctx, prID)
}
