//go:generate mockery

package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"

	"github.com/mal-as/aiqoder/internal/models"
)

var codeExtensions = map[string]bool{
	".go":    true,
	".py":    true,
	".js":    true,
	".ts":    true,
	".jsx":   true,
	".tsx":   true,
	".java":  true,
	".rs":    true,
	".c":     true,
	".cpp":   true,
	".h":     true,
	".hpp":   true,
	".cs":    true,
	".rb":    true,
	".php":   true,
	".swift": true,
	".kt":    true,
	".scala": true,
	".sh":    true,
	".yaml":  true,
	".yml":   true,
	".json":  true,
	".toml":  true,
	".sql":   true,
	".md":    true,
	".proto": true,
	".tf":    true,
}

var skipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	".idea":         true,
	".vscode":       true,
	"__pycache__":   true,
	".pytest_cache": true,
	"dist":          true,
	"build":         true,
	".cache":        true,
	".terraform":    true,
}

const maxFileSize = 512 * 1024 // 512 KB

type doc interface {
	DocumentsFromTextFile(path string) ([]*ai.Document, error)
}

type Service struct {
	doc doc
}

func New(doc doc) *Service {
	return &Service{doc: doc}
}

func (s *Service) ScanCodeFiles(rootPath string) ([]models.CodeFile, error) {
	var files []models.CodeFile

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !codeExtensions[ext] {
			return nil
		}

		info, err := d.Info()
		if err != nil || info.Size() > maxFileSize {
			return nil
		}

		docs, err := s.doc.DocumentsFromTextFile(path)
		if err != nil {
			return fmt.Errorf("doc.DocumentsFromTextFile: %w", err)
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return fmt.Errorf("filepath.Rel: %w", err)
		}

		files = append(files, models.CodeFile{
			Path: relPath,
			Docs: docs,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.WalkDir: %w", err)
	}

	return files, nil
}
