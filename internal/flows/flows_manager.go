package flows

import (
	"context"

	"github.com/mal-as/aiqoder/internal/models"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type repoStore interface {
	CreateRepository(ctx context.Context, url string) (string, error)
	InsertChunks(ctx context.Context, chunks []models.Chunk) error
}

type txManager interface {
	InTransaction(ctx context.Context, callbacks ...func(ctx context.Context) error) error
}

type gitCloner interface {
	Clone(ctx context.Context, url string) (string, error)
	Cleanup(path string)
}

type scanner interface {
	ScanCodeFiles(rootPath string) ([]models.CodeFile, error)
}

type Manager struct {
	g         *genkit.Genkit
	store     repoStore
	txm       txManager
	cloner    gitCloner
	scanner   scanner
	embedder  ai.Embedder
	model     ai.Model
	retriever ai.Retriever
	prompt    ai.Prompt
}

func NewManager(
	g *genkit.Genkit,
	store repoStore,
	txm txManager,
	cloner gitCloner,
	scanner scanner,
	embedder ai.Embedder,
	model ai.Model,
	retriever ai.Retriever,
	prompt ai.Prompt,
) *Manager {
	return &Manager{
		g:         g,
		store:     store,
		txm:       txm,
		cloner:    cloner,
		scanner:   scanner,
		embedder:  embedder,
		model:     model,
		retriever: retriever,
		prompt:    prompt,
	}
}
