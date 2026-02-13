package embeddings

import (
	"math"
	"strings"
	"sync"
	"unicode"
)

// TFIDFEngine computes TF-IDF vectors for text documents.
type TFIDFEngine struct {
	mu       sync.RWMutex
	vocab    map[string]int // term -> index in vector
	idf      map[string]float64
	vocabLen int
}

// NewTFIDFEngine creates an empty TF-IDF engine. Call BuildVocabulary before Embed.
func NewTFIDFEngine() *TFIDFEngine {
	return &TFIDFEngine{
		vocab: make(map[string]int),
		idf:   make(map[string]float64),
	}
}

// BuildVocabulary builds the vocabulary and IDF values from the corpus.
func (e *TFIDFEngine) BuildVocabulary(corpusTexts []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	stops := defaultStopWords()

	// Count document frequency for each term
	df := make(map[string]int)
	for _, text := range corpusTexts {
		seen := make(map[string]bool)
		for _, tok := range tokenize(text) {
			if stops[tok] {
				continue
			}
			if !seen[tok] {
				df[tok]++
				seen[tok] = true
			}
		}
	}

	n := float64(len(corpusTexts))
	idx := 0
	vocab := make(map[string]int, len(df))
	idf := make(map[string]float64, len(df))
	for term, count := range df {
		vocab[term] = idx
		idf[term] = math.Log(1 + n/(1+float64(count)))
		idx++
	}

	e.vocab = vocab
	e.idf = idf
	e.vocabLen = len(vocab)
}

// Embed returns the TF-IDF vector for a single text.
func (e *TFIDFEngine) Embed(text string) []float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.vocabLen == 0 {
		return nil
	}

	stops := defaultStopWords()
	tokens := tokenize(text)

	// Compute term frequencies
	tf := make(map[string]int)
	for _, tok := range tokens {
		if !stops[tok] {
			tf[tok]++
		}
	}

	total := float64(len(tokens))
	if total == 0 {
		return make([]float64, e.vocabLen)
	}

	vec := make([]float64, e.vocabLen)
	for term, count := range tf {
		if idx, ok := e.vocab[term]; ok {
			vec[idx] = (float64(count) / total) * e.idf[term]
		}
	}

	// L2 normalize
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec
}

// EmbedBatch returns TF-IDF vectors for multiple texts.
func (e *TFIDFEngine) EmbedBatch(texts []string) [][]float64 {
	results := make([][]float64, len(texts))
	for i, text := range texts {
		results[i] = e.Embed(text)
	}
	return results
}

// VocabSize returns the number of unique terms in the vocabulary.
func (e *TFIDFEngine) VocabSize() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vocabLen
}

// tokenize splits text into lowercase alpha-numeric tokens.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func defaultStopWords() map[string]bool {
	words := []string{
		"a", "an", "the", "and", "or", "but", "in", "on", "at", "to",
		"for", "of", "with", "by", "from", "is", "are", "was", "were",
		"be", "been", "being", "have", "has", "had", "do", "does", "did",
		"will", "would", "could", "should", "may", "might", "shall",
		"can", "this", "that", "these", "those", "i", "you", "he", "she",
		"it", "we", "they", "me", "him", "her", "us", "them", "my", "your",
		"his", "its", "our", "their", "what", "which", "who", "whom",
		"not", "no", "nor", "if", "then", "than", "so", "as",
	}
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}
