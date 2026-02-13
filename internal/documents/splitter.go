package documents

import (
	"fmt"
	"strings"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type Splitter struct {
	ChunkSize    int
	ChunkOverlap int
}

func NewSplitter(chunkSize, chunkOverlap int) *Splitter {
	return &Splitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

// Split splits documents into chunks.
func (s *Splitter) Split(docs []models.Document) []models.Document {
	var chunks []models.Document

	for _, doc := range docs {
		docChunks := s.splitDocument(doc)
		chunks = append(chunks, docChunks...)
	}

	return chunks
}

func (s *Splitter) splitDocument(doc models.Document) []models.Document {
	paragraphs := strings.Split(doc.Content, "\n\n")

	var chunks []models.Document
	var currentChunk strings.Builder
	chunkIndex := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if currentChunk.Len()+len(para) > s.ChunkSize && currentChunk.Len() > 0 {
			chunk := models.Document{
				ID:      fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
				Content: strings.TrimSpace(currentChunk.String()),
				Metadata: map[string]string{
					"source":      doc.Metadata["source"],
					"type":        doc.Metadata["type"],
					"category":    doc.Metadata["category"],
					"filename":    doc.Metadata["filename"],
					"chunk_index": fmt.Sprintf("%d", chunkIndex),
				},
			}
			chunks = append(chunks, chunk)
			chunkIndex++

			// Start new chunk with overlap
			currentChunk.Reset()
			if s.ChunkOverlap > 0 {
				words := strings.Fields(chunk.Content)
				overlapWords := words
				if len(words) > s.ChunkOverlap/5 {
					overlapWords = words[len(words)-s.ChunkOverlap/5:]
				}
				currentChunk.WriteString(strings.Join(overlapWords, " "))
				currentChunk.WriteString(" ")
			}
		}

		currentChunk.WriteString(para)
		currentChunk.WriteString("\n\n")
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunk := models.Document{
			ID:      fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
			Content: strings.TrimSpace(currentChunk.String()),
			Metadata: map[string]string{
				"source":      doc.Metadata["source"],
				"type":        doc.Metadata["type"],
				"category":    doc.Metadata["category"],
				"filename":    doc.Metadata["filename"],
				"chunk_index": fmt.Sprintf("%d", chunkIndex),
			},
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}
