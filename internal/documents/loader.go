package documents

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromDirectory loads all text files from a directory, skipping empty files.
func (l *Loader) LoadFromDirectory(dir string) ([]models.Document, error) {
	var documents []models.Document

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".txt" && ext != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := strings.TrimSpace(string(content))
		if text == "" {
			return nil
		}

		// Extract category from parent directory name
		category := filepath.Base(filepath.Dir(path))
		filename := filepath.Base(path)

		doc := models.Document{
			ID:      path,
			Content: text,
			Metadata: map[string]string{
				"source":   path,
				"type":     strings.TrimPrefix(ext, "."),
				"category": category,
				"filename": filename,
			},
		}

		documents = append(documents, doc)
		return nil
	})

	return documents, err
}
