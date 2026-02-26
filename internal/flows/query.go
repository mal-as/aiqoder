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

func (m *Manager) DefineQueryFlow() *core.Flow[QueryRepoInput, QueryRepoOutput, struct{}] {
	return genkit.DefineFlow(m.g, "queryRepository",
		func(ctx context.Context, input QueryRepoInput) (QueryRepoOutput, error) {
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

			resp, err := m.prompt.Execute(ctx,
				ai.WithModel(m.model),
				ai.WithInput(map[string]any{"question": input.Question}),
				ai.WithDocs(retrieved.Documents...),
			)
			if err != nil {
				return QueryRepoOutput{}, fmt.Errorf("prompt.Execute: %w", err)
			}

			return QueryRepoOutput{Answer: resp.Text()}, nil
		})
}
