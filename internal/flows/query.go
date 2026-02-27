package flows

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type QueryRepoInput struct {
	RepoID   string `json:"repoId"`
	Question string `json:"question"`
}

type QueryRepoOutput struct {
	Answer string `json:"answer"`
}

func (m *Manager) DefineQueryFlow() *core.Flow[QueryRepoInput, QueryRepoOutput, string] {
	return genkit.DefineStreamingFlow(m.g, "queryRepository",
		func(ctx context.Context, input QueryRepoInput, cb func(context.Context, string) error) (QueryRepoOutput, error) {
			if input.RepoID == "" || input.Question == "" {
				return QueryRepoOutput{}, fmt.Errorf("repoId and question are required")
			}

			retrieved, err := genkit.Retrieve(ctx, m.g,
				ai.WithRetriever(m.retriever),
				ai.WithConfig(input.RepoID),
				ai.WithTextDocs(input.Question),
			)
			if err != nil {
				return QueryRepoOutput{}, fmt.Errorf("genkit.Retrieve: %w", err)
			}

			if len(retrieved.Documents) == 0 {
				return QueryRepoOutput{
					Answer: "No relevant code was found for this repository. Make sure the repository is indexed.",
				}, nil
			}

			var modelCB func(context.Context, *ai.ModelResponseChunk) error
			if cb != nil {
				modelCB = func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
					return cb(ctx, chunk.Text())
				}
			}

			resp, err := m.prompt.Execute(ctx,
				ai.WithModel(m.model),
				ai.WithInput(map[string]any{"question": input.Question}),
				ai.WithDocs(retrieved.Documents...),
				ai.WithStreaming(modelCB),
			)
			if err != nil {
				return QueryRepoOutput{}, fmt.Errorf("prompt.Execute: %w", err)
			}

			return QueryRepoOutput{Answer: resp.Text()}, nil
		})
}
