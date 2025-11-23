package service

import (
	"context"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
)

type StatsRepository interface {
	GetStatistics(ctx context.Context) (*domain.Statistics, error)
}

type StatsService struct {
	statsRepo StatsRepository
}

func NewStatsService(statsRepo StatsRepository) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
	}
}

func (s *StatsService) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	return s.statsRepo.GetStatistics(ctx)
}
