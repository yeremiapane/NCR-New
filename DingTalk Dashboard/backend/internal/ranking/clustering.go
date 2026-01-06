package ranking

import (
	"time"

	"github.com/google/uuid"
)

// ProblemData represents a problem description with metadata
type ProblemData struct {
	ID               uuid.UUID
	DeskripsiMasalah string
	Tanggal          *time.Time
	Status           string
	Result           string
	Kategori         string
	Department       string
	Trigrams         map[string]bool // Cached trigrams
	TFIDFVector      TFIDFVector     // Cached TF-IDF vector
}

// Cluster represents a group of similar problems
type Cluster struct {
	Problems    []ProblemData
	CentroidIdx int     // Index of the centroid in Problems slice
	RPNScore    float64 // Calculated RPN score
}

// RankedProblem represents the final output for a ranked problem cluster
type RankedProblem struct {
	Rank          int      `json:"rank"`
	Description   string   `json:"description"`
	Frequency     int      `json:"frequency"`
	RPNScore      float64  `json:"rpn_score"`
	Kategori      string   `json:"kategori,omitempty"`
	SampleIDs     []string `json:"sample_ids"`
	AlgorithmInfo string   `json:"algorithm_info,omitempty"`
}

// ClusterStats contains debugging information about the clustering
type ClusterStats struct {
	TotalProblems  int      `json:"total_problems"`
	VocabularySize int      `json:"vocabulary_size"`
	ClusterCount   int      `json:"cluster_count"`
	TopTerms       []string `json:"top_terms"`
	Threshold      float64  `json:"threshold"`
	WeightTrigram  float64  `json:"weight_trigram"`
	WeightLCS      float64  `json:"weight_lcs"`
	WeightTFIDF    float64  `json:"weight_tfidf"`
}

// ClusterDescriptionsSemantic groups problems using hybrid semantic similarity
// Uses Trigram (25%) + LCS (15%) + TF-IDF (60%) for better context understanding
func ClusterDescriptionsSemantic(problems []ProblemData, threshold float64) []Cluster {
	clusters, _ := ClusterDescriptionsSemanticWithStats(problems, threshold)
	return clusters
}

// ClusterDescriptionsSemanticWithStats groups problems and returns clustering stats
func ClusterDescriptionsSemanticWithStats(problems []ProblemData, threshold float64) ([]Cluster, *ClusterStats) {
	stats := &ClusterStats{
		TotalProblems: len(problems),
		Threshold:     threshold,
	}

	if len(problems) == 0 {
		return nil, stats
	}

	// Extract all descriptions for TF-IDF fitting
	descriptions := make([]string, len(problems))
	for i, p := range problems {
		descriptions[i] = p.DeskripsiMasalah
	}

	// Initialize semantic similarity calculator
	semSim := NewSemanticSimilarity(descriptions)

	// Capture stats
	stats.VocabularySize = semSim.GetVocabularySize()
	stats.TopTerms = semSim.GetTopTerms(10)
	stats.WeightTrigram = semSim.trigramWeight
	stats.WeightLCS = semSim.lcsWeight
	stats.WeightTFIDF = semSim.tfidfWeight

	// Pre-compute trigrams and TF-IDF vectors for all problems
	for i := range problems {
		if problems[i].Trigrams == nil {
			problems[i].Trigrams = GenerateTrigrams(problems[i].DeskripsiMasalah)
		}
		problems[i].TFIDFVector = semSim.tfidfVectors[i]
	}

	// Track which problems have been assigned to a cluster
	assigned := make([]bool, len(problems))
	var clusters []Cluster

	for i := 0; i < len(problems); i++ {
		if assigned[i] {
			continue
		}

		// Start a new cluster with this problem
		cluster := Cluster{
			Problems: []ProblemData{problems[i]},
		}
		assigned[i] = true

		// Find all semantically similar problems
		for j := i + 1; j < len(problems); j++ {
			if assigned[j] {
				continue
			}

			// Calculate hybrid semantic similarity
			similarity := semSim.CalculateFromVectors(
				problems[i].DeskripsiMasalah,
				problems[j].DeskripsiMasalah,
				problems[i].Trigrams,
				problems[j].Trigrams,
				problems[i].TFIDFVector,
				problems[j].TFIDFVector,
			)

			if similarity >= threshold {
				cluster.Problems = append(cluster.Problems, problems[j])
				assigned[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	stats.ClusterCount = len(clusters)
	return clusters, stats
}

// ClusterDescriptions groups similar problem descriptions (legacy, uses only trigram)
// Kept for backwards compatibility
func ClusterDescriptions(problems []ProblemData, threshold float64) []Cluster {
	// Use semantic clustering as default now
	return ClusterDescriptionsSemantic(problems, threshold)
}

// SelectCentroidSemantic finds the most representative problem using hybrid similarity
func SelectCentroidSemantic(cluster *Cluster) {
	if len(cluster.Problems) == 0 {
		return
	}
	if len(cluster.Problems) == 1 {
		cluster.CentroidIdx = 0
		return
	}

	minAvgDistance := float64(2.0)
	centroidIdx := 0

	for i := range cluster.Problems {
		totalDistance := 0.0
		for j := range cluster.Problems {
			if i != j {
				// Use hybrid similarity for centroid selection
				trigramSim := CalculateSimilarity(cluster.Problems[i].Trigrams, cluster.Problems[j].Trigrams)
				lcsSim := CalculateLCSSimilarity(cluster.Problems[i].DeskripsiMasalah, cluster.Problems[j].DeskripsiMasalah)
				tfidfSim := CosineSimilarity(cluster.Problems[i].TFIDFVector, cluster.Problems[j].TFIDFVector)

				// Weighted combination
				similarity := (trigramSim * 0.30) + (lcsSim * 0.20) + (tfidfSim * 0.50)
				distance := 1.0 - similarity
				totalDistance += distance
			}
		}
		avgDistance := totalDistance / float64(len(cluster.Problems)-1)

		if avgDistance < minAvgDistance {
			minAvgDistance = avgDistance
			centroidIdx = i
		}
	}

	cluster.CentroidIdx = centroidIdx
}

// SelectCentroid finds the most representative problem in a cluster
// Now uses semantic similarity
func SelectCentroid(cluster *Cluster) {
	SelectCentroidSemantic(cluster)
}

// GetCentroidDescription returns the description of the cluster's centroid
func (c *Cluster) GetCentroidDescription() string {
	if len(c.Problems) == 0 {
		return ""
	}
	return c.Problems[c.CentroidIdx].DeskripsiMasalah
}

// GetClusterKeyPhrase returns a 2-4 word summary representing the cluster's problems
func (c *Cluster) GetClusterKeyPhrase(maxWords int) string {
	if len(c.Problems) == 0 {
		return ""
	}

	// Collect all descriptions in this cluster
	var descriptions []string
	for _, p := range c.Problems {
		if p.DeskripsiMasalah != "" {
			descriptions = append(descriptions, p.DeskripsiMasalah)
		}
	}

	return GetClusterSummary(descriptions, maxWords)
}

// GetSampleIDs returns up to 5 sample IDs from the cluster
func (c *Cluster) GetSampleIDs() []string {
	limit := 5
	if len(c.Problems) < limit {
		limit = len(c.Problems)
	}

	ids := make([]string, limit)
	for i := 0; i < limit; i++ {
		ids[i] = c.Problems[i].ID.String()
	}
	return ids
}

// GetMostCommonKategori finds the most frequent kategori in the cluster
func (c *Cluster) GetMostCommonKategori() string {
	if len(c.Problems) == 0 {
		return ""
	}

	kategoriCount := make(map[string]int)
	for _, p := range c.Problems {
		if p.Kategori != "" {
			kategoriCount[p.Kategori]++
		}
	}

	maxCount := 0
	mostCommon := ""
	for k, count := range kategoriCount {
		if count > maxCount {
			maxCount = count
			mostCommon = k
		}
	}

	return mostCommon
}
