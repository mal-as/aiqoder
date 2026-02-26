package repos

import (
	"context"
	"fmt"
)

func (r *Repository) CreateRepository(ctx context.Context, url string) (string, error) {
	var id string
	err := r.db.
		With(ctx).
		QueryRow(ctx, `INSERT INTO repositories (url) VALUES ($1) RETURNING id`, url).
		Scan(&id)
	if err != nil {
		return "", fmt.Errorf("db.QueryRow: %w", err)
	}

	return id, nil
}
