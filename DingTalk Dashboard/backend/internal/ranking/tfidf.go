package ranking

import (
	"math"
	"strings"
)

// TFIDFVectorizer computes TF-IDF vectors for a corpus of documents
type TFIDFVectorizer struct {
	vocabulary map[string]int // word -> index mapping
	idf        []float64      // inverse document frequencies
	docCount   int            // number of documents
}

// TFIDFVector represents a sparse TF-IDF vector
type TFIDFVector struct {
	values map[int]float64 // index -> tfidf value
	norm   float64         // L2 norm for cosine similarity
}

// NewTFIDFVectorizer creates a new TF-IDF vectorizer
func NewTFIDFVectorizer() *TFIDFVectorizer {
	return &TFIDFVectorizer{
		vocabulary: make(map[string]int),
	}
}

// Fit learns the vocabulary and IDF values from the corpus
func (v *TFIDFVectorizer) Fit(documents []string) {
	v.docCount = len(documents)
	if v.docCount == 0 {
		return
	}

	// Count document frequency for each term
	docFreq := make(map[string]int)

	for _, doc := range documents {
		// Get unique words in this document
		words := ExtractKeywords(doc)
		seen := make(map[string]bool)

		for _, word := range words {
			if !seen[word] {
				docFreq[word]++
				seen[word] = true
			}
		}
	}

	// Build vocabulary (only include words that appear in at least 2 documents)
	idx := 0
	for word, freq := range docFreq {
		if freq >= 2 {
			v.vocabulary[word] = idx
			idx++
		}
	}

	// Calculate IDF for each term
	// IDF = log(N / (1 + df)) + 1 (smoothed)
	v.idf = make([]float64, len(v.vocabulary))
	for word, i := range v.vocabulary {
		df := float64(docFreq[word])
		v.idf[i] = math.Log(float64(v.docCount)/(1+df)) + 1
	}
}

// Transform converts a document to a TF-IDF vector
func (v *TFIDFVectorizer) Transform(doc string) TFIDFVector {
	vector := TFIDFVector{
		values: make(map[int]float64),
	}

	if len(v.vocabulary) == 0 {
		return vector
	}

	// Calculate term frequencies
	words := ExtractKeywords(doc)
	tf := make(map[string]int)
	for _, word := range words {
		tf[word]++
	}

	// Calculate TF-IDF values
	totalTerms := float64(len(words))
	if totalTerms == 0 {
		return vector
	}

	sumSquares := 0.0
	for word, count := range tf {
		if idx, exists := v.vocabulary[word]; exists {
			// TF = count / total terms (normalized)
			tfValue := float64(count) / totalTerms
			// TF-IDF = TF * IDF
			tfidf := tfValue * v.idf[idx]
			vector.values[idx] = tfidf
			sumSquares += tfidf * tfidf
		}
	}

	// Calculate L2 norm
	vector.norm = math.Sqrt(sumSquares)

	return vector
}

// CosineSimilarity calculates cosine similarity between two TF-IDF vectors
// Returns value between 0 (different) and 1 (identical)
func CosineSimilarity(a, b TFIDFVector) float64 {
	if a.norm == 0 || b.norm == 0 {
		return 0
	}

	// Calculate dot product (only need to iterate over non-zero elements)
	dotProduct := 0.0
	for idx, valA := range a.values {
		if valB, exists := b.values[idx]; exists {
			dotProduct += valA * valB
		}
	}

	// Cosine similarity = dot(a,b) / (||a|| * ||b||)
	return dotProduct / (a.norm * b.norm)
}

// SemanticSimilarity calculates combined similarity using Trigram + LCS + TF-IDF
type SemanticSimilarity struct {
	vectorizer    *TFIDFVectorizer
	tfidfVectors  []TFIDFVector
	trigramWeight float64
	lcsWeight     float64
	tfidfWeight   float64
}

// NewSemanticSimilarity creates a semantic similarity calculator
func NewSemanticSimilarity(documents []string) *SemanticSimilarity {
	ss := &SemanticSimilarity{
		vectorizer:    NewTFIDFVectorizer(),
		trigramWeight: 0.25, // 25% trigram (reduced)
		lcsWeight:     0.15, // 15% LCS (reduced)
		tfidfWeight:   0.60, // 60% TF-IDF (increased for better semantic understanding)
	}

	// Fit vectorizer on all documents
	ss.vectorizer.Fit(documents)

	// Pre-compute TF-IDF vectors for all documents
	ss.tfidfVectors = make([]TFIDFVector, len(documents))
	for i, doc := range documents {
		ss.tfidfVectors[i] = ss.vectorizer.Transform(doc)
	}

	return ss
}

// Calculate computes hybrid similarity between two documents
func (ss *SemanticSimilarity) Calculate(idx1, idx2 int, text1, text2 string, trigrams1, trigrams2 map[string]bool) float64 {
	// Trigram similarity
	trigramSim := CalculateSimilarity(trigrams1, trigrams2)

	// LCS similarity
	lcsSim := CalculateLCSSimilarity(text1, text2)

	// TF-IDF cosine similarity
	var tfidfSim float64
	if idx1 < len(ss.tfidfVectors) && idx2 < len(ss.tfidfVectors) {
		tfidfSim = CosineSimilarity(ss.tfidfVectors[idx1], ss.tfidfVectors[idx2])
	}

	// Weighted combination
	combined := (trigramSim * ss.trigramWeight) +
		(lcsSim * ss.lcsWeight) +
		(tfidfSim * ss.tfidfWeight)

	return combined
}

// CalculateFromVectors computes similarity when you have pre-computed vectors
func (ss *SemanticSimilarity) CalculateFromVectors(text1, text2 string, trigrams1, trigrams2 map[string]bool, vec1, vec2 TFIDFVector) float64 {
	trigramSim := CalculateSimilarity(trigrams1, trigrams2)
	lcsSim := CalculateLCSSimilarity(text1, text2)
	tfidfSim := CosineSimilarity(vec1, vec2)

	return (trigramSim * ss.trigramWeight) +
		(lcsSim * ss.lcsWeight) +
		(tfidfSim * ss.tfidfWeight)
}

// GetVocabularySize returns the number of unique terms learned
func (ss *SemanticSimilarity) GetVocabularySize() int {
	return len(ss.vectorizer.vocabulary)
}

// GetTopTerms returns the top N terms by IDF (most distinctive/important)
func (ss *SemanticSimilarity) GetTopTerms(n int) []string {
	type termIDF struct {
		term string
		idf  float64
	}

	// Collect all terms with their IDF
	terms := make([]termIDF, 0, len(ss.vectorizer.vocabulary))
	for term, idx := range ss.vectorizer.vocabulary {
		terms = append(terms, termIDF{term: term, idf: ss.vectorizer.idf[idx]})
	}

	// Sort by IDF descending (bubble sort for simplicity)
	for i := 0; i < len(terms)-1; i++ {
		for j := 0; j < len(terms)-i-1; j++ {
			if terms[j].idf < terms[j+1].idf {
				terms[j], terms[j+1] = terms[j+1], terms[j]
			}
		}
	}

	// Return top N
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(terms); i++ {
		result = append(result, terms[i].term)
	}

	return result
}

// NormalizeText cleans and normalizes text for better matching
func NormalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove extra whitespace
	text = normalizeWhitespace(text)

	return text
}
