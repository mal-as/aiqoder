package scanner_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mal-as/aiqoder/internal/services/scanner"
	"github.com/mal-as/aiqoder/internal/services/scanner/mocks"
)

func TestService_ScanCodeFiles_Success(t *testing.T) {
	rootDir := t.TempDir()

	goFile := filepath.Join(rootDir, "main.go")
	pyFile := filepath.Join(rootDir, "script.py")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0o600))
	require.NoError(t, os.WriteFile(pyFile, []byte("print('hello')"), 0o600))

	goDocs := []*ai.Document{ai.DocumentFromText("package main", nil)}
	pyDocs := []*ai.Document{ai.DocumentFromText("print('hello')", nil)}

	mockDoc := mocks.NewMockdoc(t)
	mockDoc.EXPECT().DocumentsFromTextFile(goFile).Return(goDocs, nil)
	mockDoc.EXPECT().DocumentsFromTextFile(pyFile).Return(pyDocs, nil)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestService_ScanCodeFiles_SkipsUnsupportedExtensions(t *testing.T) {
	rootDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "file.xyz"), []byte("data"), 0o600))

	goFile := filepath.Join(rootDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0o600))

	mockDoc := mocks.NewMockdoc(t)
	mockDoc.EXPECT().DocumentsFromTextFile(goFile).Return([]*ai.Document{}, nil)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestService_ScanCodeFiles_SkipsSkippedDirs(t *testing.T) {
	rootDir := t.TempDir()

	// Place a .go file inside .git — it must be skipped
	gitDir := filepath.Join(rootDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "hook.go"), []byte("package git"), 0o600))

	goFile := filepath.Join(rootDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0o600))

	mockDoc := mocks.NewMockdoc(t)
	mockDoc.EXPECT().DocumentsFromTextFile(goFile).Return([]*ai.Document{}, nil)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestService_ScanCodeFiles_DocumentsError(t *testing.T) {
	rootDir := t.TempDir()

	goFile := filepath.Join(rootDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0o600))

	docErr := errors.New("read error")
	mockDoc := mocks.NewMockdoc(t)
	mockDoc.EXPECT().DocumentsFromTextFile(goFile).Return(nil, docErr)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.Error(t, err)
	assert.Nil(t, files)
	require.ErrorContains(t, err, "doc.DocumentsFromTextFile")
	assert.ErrorIs(t, err, docErr)
}

func TestService_ScanCodeFiles_PathIsTrimmedToRelative(t *testing.T) {
	rootDir := t.TempDir()

	subDir := filepath.Join(rootDir, "cmd")
	require.NoError(t, os.Mkdir(subDir, 0o750))

	goFile := filepath.Join(subDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0o600))

	docs := []*ai.Document{ai.DocumentFromText("package main", nil)}
	mockDoc := mocks.NewMockdoc(t)
	mockDoc.EXPECT().DocumentsFromTextFile(goFile).Return(docs, nil)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "cmd/main.go", files[0].Path)
}

func TestService_ScanCodeFiles_EmptyDirectory(t *testing.T) {
	rootDir := t.TempDir()

	mockDoc := mocks.NewMockdoc(t)

	svc := scanner.New(mockDoc)
	files, err := svc.ScanCodeFiles(rootDir)

	require.NoError(t, err)
	assert.Empty(t, files)
}
