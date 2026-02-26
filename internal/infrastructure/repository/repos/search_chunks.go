package repos

import (
	"context"
	"fmt"

	"github.com/mal-as/aiqoder/internal/models"

	"github.com/pgvector/pgvector-go"
)

func (r *Repository) SearchChunks(ctx context.Context, repoID string, embedding []float32, limit int) ([]models.ChunkResult, error) {
	q := `
		SELECT file_path, content
		FROM   code_chunks
		WHERE  repo_id = $1
		ORDER  BY embedding <=> $2
		LIMIT  $3`

	rows, err := r.db.With(ctx).Query(ctx, q, repoID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var results []models.ChunkResult
	for rows.Next() {
		var res models.ChunkResult

		if err = rows.Scan(&res.FilePath, &res.Content); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		results = append(results, res)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return results, nil
}
