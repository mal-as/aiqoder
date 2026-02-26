//go:generate mockery

package documents

import (
	"fmt"
	"os"

	"github.com/firebase/genkit/go/ai"
)

type splitter interface {
	SplitText(text string) ([]string, error)
}

type Service struct {
	splitter splitter
}

func NewService(splitter splitter) *Service {
	return &Service{splitter: splitter}
}

func (s *Service) DocumentsFromTextFile(path string) ([]*ai.Document, error) {
	rawData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	chunks, err := s.splitter.SplitText(string(rawData))
	if err != nil {
		return nil, fmt.Errorf("splitter.SplitText: %w", err)
	}

	docs := make([]*ai.Document, 0, len(chunks))
	for _, chunk := range chunks {
		docs = append(docs, ai.DocumentFromText(chunk, nil))
	}

	return docs, nil
}
