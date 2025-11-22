package service

import (
	"context"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type TeamService struct {
	teamRepo TeamRepository
}

func NewTeamService(teamRepo TeamRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, teamName string, members []*domain.User) (*domain.Team, error) {
	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.ErrTeamExists(teamName)
	}

	team := domain.NewTeam(teamName, members)

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.teamRepo.GetByName(ctx, teamName)
}
