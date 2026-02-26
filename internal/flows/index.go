package flows

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/mal-as/aiqoder/internal/models"
)

type IndexRepoInput struct {
	RepoURL string `json:"repoUrl"`
}

type IndexRepoOutput struct {
	RepoID string `json:"repoId"`
}

func (m *Manager) DefineIndexFlow() *core.Flow[IndexRepoInput, IndexRepoOutput, struct{}] {
	return genkit.DefineFlow(m.g, "indexRepository",
		func(ctx context.Context, input IndexRepoInput) (IndexRepoOutput, error) {
			if input.RepoURL == "" {
				return IndexRepoOutput{}, fmt.Errorf("repoUrl is required")
			}

			codeFiles, err := m.codeFiles(ctx, &input)
			if err != nil {
				return IndexRepoOutput{}, err
			}

			chunks, err := m.chunkCodeFiles(ctx, codeFiles)
			if err != nil {
				return IndexRepoOutput{}, err
			}

			repoID, err := m.persistChunks(ctx, &input, chunks)
			if err != nil {
				return IndexRepoOutput{}, err
			}

			slog.InfoContext(ctx, "repository indexed",
				slog.String("url", input.RepoURL),
				slog.String("repo_id", repoID),
				slog.Int("files", len(codeFiles)),
				slog.Int("chunks", len(chunks)),
			)

			return IndexRepoOutput{RepoID: repoID}, nil
		})
}

func (m *Manager) codeFiles(ctx context.Context, input *IndexRepoInput) ([]models.CodeFile, error) {
	repoPath, err := m.cloner.Clone(ctx, input.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("cloner.Clone: %w", err)
	}
	defer m.cloner.Cleanup(repoPath)

	codeFiles, err := m.scanner.ScanCodeFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("scanner.ScanCodeFiles: %w", err)
	}

	if len(codeFiles) == 0 {
		return nil, fmt.Errorf("no indexable code files found in repository")
	}

	return codeFiles, nil
}

func (m *Manager) chunkCodeFiles(ctx context.Context, codeFiles []models.CodeFile) ([]models.Chunk, error) {
	chunks := make([]models.Chunk, 0, len(codeFiles))

	for _, codeFile := range codeFiles {
		embedResp, err := genkit.Embed(ctx, m.g,
			ai.WithEmbedder(m.embedder),
			ai.WithDocs(codeFile.Docs...),
		)
		if err != nil {
			return nil, fmt.Errorf("genkit.Embed: %w", err)
		}

		if len(embedResp.Embeddings) == 0 {
			continue
		}

		if len(embedResp.Embeddings) != len(codeFile.Docs) {
			return nil, fmt.Errorf("embeddings count mismatch: got %d embeddings for %d docs in %s",
				len(embedResp.Embeddings), len(codeFile.Docs), codeFile.Path)
		}

		for i, doc := range codeFile.Docs {
			chunks = append(chunks, models.Chunk{
				FilePath:  codeFile.Path,
				Content:   docContent(doc),
				Embedding: embedResp.Embeddings[i].Embedding,
			})
		}
	}

	return chunks, nil
}

func (m *Manager) persistChunks(ctx context.Context, input *IndexRepoInput, chunks []models.Chunk) (string, error) {
	var repoID string

	err := m.txm.InTransaction(ctx, func(ctx context.Context) error {
		var err error

		repoID, err = m.store.CreateRepository(ctx, input.RepoURL)
		if err != nil {
			return fmt.Errorf("store.CreateRepository: %w", err)
		}

		for i := range chunks {
			chunks[i].RepoID = repoID
		}

		if err = m.store.InsertChunks(ctx, chunks); err != nil {
			return fmt.Errorf("store.InsertChunks: %w", err)
		}

		return nil
	})

	return repoID, err
}

func docContent(doc *ai.Document) string {
	parts := make([]string, 0, len(doc.Content))
	for _, part := range doc.Content {
		parts = append(parts, part.Text)
	}

	return strings.Join(parts, "\n")
}
