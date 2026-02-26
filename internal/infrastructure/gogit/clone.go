package gogit

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	gogit "github.com/go-git/go-git/v6"
)

type cloneFunc func(ctx context.Context, path string, o *gogit.CloneOptions) (*gogit.Repository, error)

type Cloner struct {
	cloneFunc cloneFunc
}

func New() *Cloner {
	return &Cloner{
		cloneFunc: gogit.PlainCloneContext,
	}
}

func (g *Cloner) Clone(ctx context.Context, url string) (string, error) {
	dir, err := os.MkdirTemp("", "repo_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return "", fmt.Errorf("os.MkdirTemp: %w", err)
	}

	_, err = g.cloneFunc(ctx, dir, &gogit.CloneOptions{
		URL:          url,
		Depth:        1,
		SingleBranch: true,
	})
	if err != nil {
		_ = os.RemoveAll(dir) // best-effort cleanup on clone failure
		return "", fmt.Errorf("gogit.PlainCloneContext: %w", err)
	}

	return dir, nil
}

func (g *Cloner) Cleanup(path string) {
	if path != "" {
		_ = os.RemoveAll(path)
	}
}
