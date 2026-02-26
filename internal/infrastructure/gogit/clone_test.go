package gogit

import (
	"context"
	"errors"
	"os"
	"testing"

	gogit "github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloner_Clone_Success(t *testing.T) {
	var capturedPath string
	var capturedURL string

	cloner := &Cloner{
		cloneFunc: func(_ context.Context, path string, o *gogit.CloneOptions) (*gogit.Repository, error) {
			capturedPath = path
			capturedURL = o.URL
			return nil, nil
		},
	}

	dir, err := cloner.Clone(context.Background(), "https://example.com/repo.git")
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Equal(t, dir, capturedPath)
	assert.Equal(t, "https://example.com/repo.git", capturedURL)

	t.Cleanup(func() { _ = os.RemoveAll(dir) })
}

func TestCloner_Clone_PassesCorrectOptions(t *testing.T) {
	var capturedOpts *gogit.CloneOptions

	cloner := &Cloner{
		cloneFunc: func(_ context.Context, _ string, o *gogit.CloneOptions) (*gogit.Repository, error) {
			capturedOpts = o
			return nil, nil
		},
	}

	dir, err := cloner.Clone(context.Background(), "https://example.com/repo.git")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	require.NotNil(t, capturedOpts)
	assert.Equal(t, 1, capturedOpts.Depth)
	assert.True(t, capturedOpts.SingleBranch)
}

func TestCloner_Clone_CloneFuncError(t *testing.T) {
	cloneErr := errors.New("connection refused")

	cloner := &Cloner{
		cloneFunc: func(_ context.Context, _ string, _ *gogit.CloneOptions) (*gogit.Repository, error) {
			return nil, cloneErr
		},
	}

	dir, err := cloner.Clone(context.Background(), "https://example.com/repo.git")
	require.Error(t, err)
	assert.Empty(t, dir)
	require.ErrorContains(t, err, "gogit.PlainCloneContext")
	assert.ErrorIs(t, err, cloneErr)
}

func TestCloner_Cleanup_RemovesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_cleanup_")
	require.NoError(t, err)

	cloner := New()
	cloner.Cleanup(tmpDir)

	_, statErr := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(statErr))
}

func TestCloner_Cleanup_EmptyPath(_ *testing.T) {
	cloner := New()
	// Must not panic on empty path
	cloner.Cleanup("")
}
