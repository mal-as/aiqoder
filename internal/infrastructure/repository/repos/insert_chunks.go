package repos

import (
	"context"
	"fmt"

	"github.com/mal-as/aiqoder/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
)

func (r *Repository) InsertChunks(ctx context.Context, chunks []models.Chunk) (retErr error) {
	if len(chunks) == 0 {
		return nil
	}

	q := `INSERT INTO code_chunks (repo_id, file_path, content, embedding) VALUES ($1, $2, $3, $4)`
	batch := &pgx.Batch{}

	for _, c := range chunks {
		batch.Queue(q, c.RepoID, c.FilePath, c.Content, pgvector.NewVector(c.Embedding))
	}

	if err := r.db.With(ctx).SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("db.SendBatch: %w", err)
	}

	return nil
}
