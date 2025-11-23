package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
)

type TeamRepository struct {
	db *DB
}

func NewTeamRepository(db *DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	teamQuery := `
		INSERT INTO teams (team_name, created_at)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, teamQuery, team.TeamName, team.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	userQuery := `
		INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username,
		    team_name = EXCLUDED.team_name,
		    is_active = EXCLUDED.is_active,
		    updated_at = EXCLUDED.updated_at
	`

	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, userQuery,
			member.UserID,
			member.Username,
			member.TeamName,
			member.IsActive,
			member.CreatedAt,
			member.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create/update user %s: %w", member.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	teamQuery := `
		SELECT team_name, created_at
		FROM teams
		WHERE team_name = $1
	`

	team := &domain.Team{}
	err := r.db.QueryRowContext(ctx, teamQuery, teamName).Scan(
		&team.TeamName,
		&team.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrTeamNotFound(teamName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	userQuery := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, userQuery, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	members := make([]*domain.User, 0)
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
		members = append(members, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	team.Members = members
	return team, nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}
