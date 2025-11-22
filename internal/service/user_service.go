package service

import (
	"context"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
	GetActiveByTeamExcluding(ctx context.Context, teamName, excludeUserID string) ([]*domain.User, error)
}

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if err := s.userRepo.SetActive(ctx, userID, isActive); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
