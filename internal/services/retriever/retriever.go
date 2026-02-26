package retriever

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/mal-as/aiqoder/internal/models"
)

type chunkSearcher interface {
	SearchChunks(ctx context.Context, repoID string, embedding []float32, limit int) ([]models.ChunkResult, error)
}

const defaultTopK = 10

func Define(g *genkit.Genkit, store chunkSearcher, embedder ai.Embedder) ai.Retriever {
	return genkit.DefineRetriever(g, "pgvector/code-chunks", nil,
		func(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
			repoID, ok := req.Options.(string)
			if !ok || repoID == "" {
				return nil, fmt.Errorf("retriever options must be a non-empty repository ID string")
			}

			eres, err := genkit.Embed(ctx, g,
				ai.WithEmbedder(embedder),
				ai.WithDocs(req.Query),
			)
			if err != nil {
				return nil, fmt.Errorf("embed query: %w", err)
			}
			if len(eres.Embeddings) == 0 {
				return nil, fmt.Errorf("no embeddings returned for query")
			}

			results, err := store.SearchChunks(ctx, repoID, eres.Embeddings[0].Embedding, defaultTopK)
			if err != nil {
				return nil, err
			}

			res := ai.RetrieverResponse{
				Documents: make([]*ai.Document, 0, len(results)),
			}
			for _, r := range results {
				res.Documents = append(res.Documents, &ai.Document{
					Content:  []*ai.Part{ai.NewTextPart("// File: " + r.FilePath + "\n\n" + r.Content)},
					Metadata: map[string]any{"file_path": r.FilePath},
				})
			}

			return &res, nil
		},
	)
}
