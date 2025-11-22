package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, team_name = $3, is_active = $4, updated_at = $5
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.ErrUserNotFound(user.UserID)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound(userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by team: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetActiveByTeamExcluding(ctx context.Context, teamName, excludeUserID string) ([]*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users
		SET is_active = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, isActive)
	if err != nil {
		return fmt.Errorf("failed to set user active status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.ErrUserNotFound(userID)
	}

	return nil
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}
