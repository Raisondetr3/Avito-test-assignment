package service

import (
	"context"
	"fmt"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/repository/postgres"
)

type BulkDeactivationService struct {
	userRepo         UserRepository
	teamRepo         TeamRepository
	prRepo           PRRepository
	reviewerAssigner *ReviewerAssigner
}

func NewBulkDeactivationService(
	userRepo UserRepository,
	teamRepo TeamRepository,
	prRepo PRRepository,
	reviewerAssigner *ReviewerAssigner,
) *BulkDeactivationService {
	return &BulkDeactivationService{
		userRepo:         userRepo,
		teamRepo:         teamRepo,
		prRepo:           prRepo,
		reviewerAssigner: reviewerAssigner,
	}
}

//nolint:gocyclo
func (s *BulkDeactivationService) DeactivateUsersAndReassignPRs(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (*domain.BulkDeactivationResult, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	usersToDeactivate := userIDs
	if len(userIDs) == 0 {
		teamUsers, err := s.userRepo.GetByTeam(ctx, teamName)
		if err != nil {
			return nil, fmt.Errorf("failed to get team users: %w", err)
		}
		usersToDeactivate = make([]string, 0, len(teamUsers))
		for _, user := range teamUsers {
			usersToDeactivate = append(usersToDeactivate, user.UserID)
		}
	} else {
		users, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}
		for _, user := range users {
			if user.TeamName != team.TeamName {
				return nil, fmt.Errorf("user %s is not in team %s", user.UserID, teamName)
			}
		}
	}

	if len(usersToDeactivate) == 0 {
		return &domain.BulkDeactivationResult{
			DeactivatedUsers: []string{},
			ReassignedPRs:    []domain.ReassignedPR{},
			SkippedPRs:       []domain.SkippedPR{},
		}, nil
	}

	openPRsInfo, err := s.prRepo.(*postgres.PRRepository).GetOpenPRsWithReviewers(ctx, usersToDeactivate) //nolint:errcheck
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs: %w", err)
	}

	err = s.userRepo.BulkDeactivate(ctx, usersToDeactivate)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk deactivate users: %w", err)
	}

	reassignedPRs := make([]domain.ReassignedPR, 0)
	skippedPRs := make([]domain.SkippedPR, 0)

	for _, prInfo := range openPRsInfo {
		candidates, err := s.userRepo.GetActiveByTeamExcluding(ctx, prInfo.ReviewerTeam, prInfo.ReviewerID)
		if err != nil {
			skippedPRs = append(skippedPRs, domain.SkippedPR{
				PullRequestID: prInfo.PullRequestID,
				Reason:        fmt.Sprintf("failed to get candidates: %v", err),
			})
			continue
		}

		excludeUsers := []string{prInfo.ReviewerID}
		pr, err := s.prRepo.GetByID(ctx, prInfo.PullRequestID)
		if err != nil {
			skippedPRs = append(skippedPRs, domain.SkippedPR{
				PullRequestID: prInfo.PullRequestID,
				Reason:        fmt.Sprintf("failed to get PR: %v", err),
			})
			continue
		}

		excludeUsers = append(excludeUsers, pr.AssignedReviewers...)

		availableCandidates := make([]*domain.User, 0)
		for _, candidate := range candidates {
			isExcluded := false
			for _, excludeID := range excludeUsers {
				if candidate.UserID == excludeID {
					isExcluded = true
					break
				}
			}
			if !isExcluded {
				availableCandidates = append(availableCandidates, candidate)
			}
		}

		if len(availableCandidates) == 0 {
			skippedPRs = append(skippedPRs, domain.SkippedPR{
				PullRequestID: prInfo.PullRequestID,
				Reason:        "no active replacement candidate in team",
			})
			continue
		}

		newReviewerID, ok := s.reviewerAssigner.SelectRandomReviewer(availableCandidates)
		if !ok {
			skippedPRs = append(skippedPRs, domain.SkippedPR{
				PullRequestID: prInfo.PullRequestID,
				Reason:        "failed to select reviewer",
			})
			continue
		}

		err = s.prRepo.ReplaceReviewer(ctx, prInfo.PullRequestID, prInfo.ReviewerID, newReviewerID)
		if err != nil {
			skippedPRs = append(skippedPRs, domain.SkippedPR{
				PullRequestID: prInfo.PullRequestID,
				Reason:        fmt.Sprintf("failed to replace reviewer: %v", err),
			})
			continue
		}

		reassignedPRs = append(reassignedPRs, domain.ReassignedPR{
			PullRequestID: prInfo.PullRequestID,
			OldReviewerID: prInfo.ReviewerID,
			NewReviewerID: newReviewerID,
		})
	}

	return &domain.BulkDeactivationResult{
		DeactivatedUsers: usersToDeactivate,
		ReassignedPRs:    reassignedPRs,
		SkippedPRs:       skippedPRs,
	}, nil
}
