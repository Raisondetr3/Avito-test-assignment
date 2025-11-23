package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
)

type PRRepository struct {
	db *DB
}

func NewPRRepository(db *DB) *PRRepository {
	return &PRRepository{db: db}
}

func (r *PRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, prQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
		pr.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerQuery := `
			INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
		`

		for _, reviewerID := range pr.AssignedReviewers {
			_, err = tx.ExecContext(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
			if err != nil {
				return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	prQuery := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, updated_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	pr := &domain.PullRequest{}
	err := r.db.QueryRowContext(ctx, prQuery, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.UpdatedAt,
		&pr.MergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrPRNotFound(prID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	reviewerQuery := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at
	`

	rows, err := r.db.QueryContext(ctx, reviewerQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	reviewers := make([]string, 0, 2)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}

	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (r *PRRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	query := `
		UPDATE pull_requests
		SET pull_request_name = $2, status = $3, updated_at = $4, merged_at = $5
		WHERE pull_request_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.Status,
		pr.UpdatedAt,
		pr.MergedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.ErrPRNotFound(pr.PullRequestID)
	}

	return nil
}

func (r *PRRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	query := `
		UPDATE pr_reviewers
		SET reviewer_id = $3, assigned_at = CURRENT_TIMESTAMP
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to replace reviewer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.ErrNotAssigned(oldReviewerID, prID)
	}

	return nil
}

func (r *PRRepository) GetByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error) {
	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	defer rows.Close()

	prs := make([]*domain.PullRequest, 0)
	for rows.Next() {
		pr := &domain.PullRequest{}
		var createdAt any
		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull requests: %w", err)
	}

	return prs, nil
}

func (r *PRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return exists, nil
}

type PRReviewerInfo struct {
	PullRequestID string
	ReviewerID    string
	ReviewerTeam  string
}

func (r *PRRepository) GetOpenPRsWithReviewers(ctx context.Context, reviewerIDs []string) ([]PRReviewerInfo, error) {
	if len(reviewerIDs) == 0 {
		return []PRReviewerInfo{}, nil
	}

	query := `
		SELECT pr.pull_request_id, prr.reviewer_id, u.team_name
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		INNER JOIN users u ON prr.reviewer_id = u.user_id
		WHERE pr.status = 'OPEN' AND prr.reviewer_id = ANY($1)
		ORDER BY pr.created_at
	`

	rows, err := r.db.QueryContext(ctx, query, reviewerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs with reviewers: %w", err)
	}
	defer rows.Close()

	infos := make([]PRReviewerInfo, 0)
	for rows.Next() {
		var info PRReviewerInfo
		err := rows.Scan(&info.PullRequestID, &info.ReviewerID, &info.ReviewerTeam)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR reviewer info: %w", err)
		}
		infos = append(infos, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PR reviewer infos: %w", err)
	}

	return infos, nil
}
