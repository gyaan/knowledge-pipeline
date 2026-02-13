package embeddings

// Embedder is the interface for generating text embeddings.
type Embedder interface {
	Embed(text string) []float64
	EmbedBatch(texts []string) [][]float64
	VocabSize() int
}
