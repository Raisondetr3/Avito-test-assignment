package postgres

import (
	"context"
	"fmt"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
)

type StatsRepository struct {
	db *DB
}

func NewStatsRepository(db *DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	stats := &domain.Statistics{}

	prStats, err := r.getPullRequestStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}
	stats.PullRequests = prStats

	userStats, err := r.getUserStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	stats.Users = userStats

	teamStats, err := r.getTeamStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats: %w", err)
	}
	stats.Teams = teamStats

	topReviewers, err := r.getTopReviewers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top reviewers: %w", err)
	}
	stats.TopReviewers = topReviewers

	return stats, nil
}

func (r *StatsRepository) getPullRequestStats(ctx context.Context) (domain.PullRequestStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'OPEN') as open,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged
		FROM pull_requests
	`

	var stats domain.PullRequestStats
	err := r.db.QueryRowContext(ctx, query).Scan(&stats.Total, &stats.Open, &stats.Merged)
	if err != nil {
		return domain.PullRequestStats{}, err
	}

	return stats, nil
}

func (r *StatsRepository) getUserStats(ctx context.Context) (domain.UserStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active,
			COUNT(*) FILTER (WHERE is_active = false) as inactive
		FROM users
	`

	var stats domain.UserStats
	err := r.db.QueryRowContext(ctx, query).Scan(&stats.Total, &stats.Active, &stats.Inactive)
	if err != nil {
		return domain.UserStats{}, err
	}

	return stats, nil
}

func (r *StatsRepository) getTeamStats(ctx context.Context) (domain.TeamStats, error) {
	query := `SELECT COUNT(DISTINCT team_name) FROM users WHERE team_name IS NOT NULL`

	var stats domain.TeamStats
	err := r.db.QueryRowContext(ctx, query).Scan(&stats.Total)
	if err != nil {
		return domain.TeamStats{}, err
	}

	return stats, nil
}

func (r *StatsRepository) getTopReviewers(ctx context.Context) ([]domain.ReviewerStat, error) {
	query := `
		SELECT
			u.user_id,
			u.username,
			COUNT(DISTINCT prr.pull_request_id) as review_count
		FROM users u
		LEFT JOIN pr_reviewers prr ON u.user_id = prr.reviewer_id
		GROUP BY u.user_id, u.username
		HAVING COUNT(DISTINCT prr.pull_request_id) > 0
		ORDER BY review_count DESC
		LIMIT 10
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]domain.ReviewerStat, 0)
	for rows.Next() {
		var reviewer domain.ReviewerStat
		err := rows.Scan(&reviewer.UserID, &reviewer.Username, &reviewer.ReviewCount)
		if err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}
