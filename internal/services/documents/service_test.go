package documents

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mal-as/aiqoder/internal/services/documents/mocks"
)

func TestService_DocumentsFromTextFile_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	content := "hello world content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	chunks := []string{"hello world", "content"}
	mockSplitter := mocks.NewMocksplitter(t)
	mockSplitter.EXPECT().SplitText(content).Return(chunks, nil)

	svc := NewService(mockSplitter)
	docs, err := svc.DocumentsFromTextFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Len(t, docs, len(chunks))
}

func TestService_DocumentsFromTextFile_ReadError(t *testing.T) {
	mockSplitter := mocks.NewMocksplitter(t)

	svc := NewService(mockSplitter)
	docs, err := svc.DocumentsFromTextFile("/nonexistent/path/file.txt")

	require.Error(t, err)
	assert.Nil(t, docs)
	assert.ErrorContains(t, err, "os.ReadFile")
}

func TestService_DocumentsFromTextFile_SplitError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	content := "some content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	splitErr := errors.New("split failed")
	mockSplitter := mocks.NewMocksplitter(t)
	mockSplitter.EXPECT().SplitText(content).Return(nil, splitErr)

	svc := NewService(mockSplitter)
	docs, err := svc.DocumentsFromTextFile(tmpFile.Name())

	require.Error(t, err)
	assert.Nil(t, docs)
	require.ErrorContains(t, err, "splitter.SplitText")
	assert.ErrorIs(t, err, splitErr)
}

func TestService_DocumentsFromTextFile_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })
	require.NoError(t, tmpFile.Close())

	mockSplitter := mocks.NewMocksplitter(t)
	mockSplitter.EXPECT().SplitText("").Return([]string{}, nil)

	svc := NewService(mockSplitter)
	docs, err := svc.DocumentsFromTextFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Empty(t, docs)
}
