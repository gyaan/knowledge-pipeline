package main

import (
	"flag"
	"log"

	"github.com/gyaan/knowledge-pipeline/internal/config"
	"github.com/gyaan/knowledge-pipeline/internal/documents"
	"github.com/gyaan/knowledge-pipeline/internal/embeddings"
)

func main() {
	docsPath := flag.String("docs", "", "Path to documents directory (overrides config)")
	flag.Parse()

	log.Println("Starting document ingestion (dry run)...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	path := cfg.KnowledgeBasePath
	if *docsPath != "" {
		path = *docsPath
	}

	// Load documents
	loader := documents.NewLoader()
	log.Printf("Loading documents from %s...", path)
	docs, err := loader.LoadFromDirectory(path)
	if err != nil {
		log.Fatalf("Failed to load documents: %v", err)
	}
	log.Printf("Loaded %d documents", len(docs))

	for _, doc := range docs {
		log.Printf("  - %s (%s/%s, %d chars)",
			doc.Metadata["filename"],
			doc.Metadata["category"],
			doc.Metadata["type"],
			len(doc.Content))
	}

	// Split into chunks
	splitter := documents.NewSplitter(cfg.ChunkSize, cfg.ChunkOverlap)
	chunks := splitter.Split(docs)
	log.Printf("Created %d chunks (size=%d, overlap=%d)", len(chunks), cfg.ChunkSize, cfg.ChunkOverlap)

	// Build TF-IDF vocabulary
	tfidf := embeddings.NewTFIDFEngine()
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}
	tfidf.BuildVocabulary(texts)
	log.Printf("Built TF-IDF vocabulary: %d unique terms", tfidf.VocabSize())

	// Embed all chunks
	vecs := tfidf.EmbedBatch(texts)

	nonZero := 0
	for _, v := range vecs {
		for _, val := range v {
			if val != 0 {
				nonZero++
				break
			}
		}
	}

	log.Printf("Generated %d embeddings (%d non-zero vectors, dim=%d)",
		len(vecs), nonZero, tfidf.VocabSize())
	log.Println("Dry run complete. Use cmd/server to run with live ingestion.")
}
