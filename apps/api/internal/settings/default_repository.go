package settings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/real-staging-ai/api/internal/storage"
)

// DefaultRepository implements Repository using PostgreSQL.
type DefaultRepository struct {
	db storage.PgxPool
}

// Ensure DefaultRepository implements Repository.
var _ Repository = (*DefaultRepository)(nil)

// NewDefaultRepository creates a new DefaultRepository.
func NewDefaultRepository(db storage.PgxPool) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// GetByKey retrieves a setting by its key.
func (r *DefaultRepository) GetByKey(ctx context.Context, key string) (*Setting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM settings
		WHERE key = $1
	`

	var setting Setting
	var updatedBy *string

	err := r.db.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
		&updatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("setting not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	setting.UpdatedBy = updatedBy

	return &setting, nil
}

// Update updates a setting value.
func (r *DefaultRepository) Update(ctx context.Context, key, value, userID string) error {
	query := `
		UPDATE settings
		SET value = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
	`

	result, err := r.db.Exec(ctx, query, value, userID, key)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}

	return nil
}

// List retrieves all settings.
func (r *DefaultRepository) List(ctx context.Context) ([]Setting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM settings
		ORDER BY key
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var setting Setting
		var updatedBy *string

		err := rows.Scan(
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.UpdatedAt,
			&updatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		setting.UpdatedBy = updatedBy
		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}
