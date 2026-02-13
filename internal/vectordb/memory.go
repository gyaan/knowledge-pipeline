package vectordb

import (
	"sort"
	"sync"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type MemoryVectorDB struct {
	documents []models.Document
	mu        sync.RWMutex
}

func NewMemoryVectorDB() *MemoryVectorDB {
	return &MemoryVectorDB{
		documents: make([]models.Document, 0),
	}
}

// Add adds documents to the vector database.
func (db *MemoryVectorDB) Add(docs []models.Document) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.documents = append(db.documents, docs...)
}

// Search performs cosine similarity search and returns top K results.
func (db *MemoryVectorDB) Search(queryEmbedding []float64, topK int) []models.SourceDocument {
	db.mu.RLock()
	defer db.mu.RUnlock()

	type scoredDoc struct {
		doc   models.Document
		score float64
	}

	scores := make([]scoredDoc, 0, len(db.documents))

	for _, doc := range db.documents {
		if len(doc.Embedding) == 0 {
			continue
		}

		score := CosineSimilarity(queryEmbedding, doc.Embedding)
		scores = append(scores, scoredDoc{doc: doc, score: score})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]models.SourceDocument, topK)
	for i := 0; i < topK; i++ {
		results[i] = models.SourceDocument{
			Content:  scores[i].doc.Content,
			Metadata: scores[i].doc.Metadata,
			Score:    scores[i].score,
		}
	}

	return results
}

// Count returns the number of documents.
func (db *MemoryVectorDB) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.documents)
}
