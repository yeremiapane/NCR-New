package ranking

import (
	"math"
	"time"
)

// RPNConfig holds configuration for RPN calculation
type RPNConfig struct {
	FrequencyWeight float64 // Weight for frequency component
	RecencyWeight   float64 // Weight for recency component
	RecencyDays     int     // Number of days to consider for recency decay
}

// DefaultRPNConfig returns default RPN configuration
func DefaultRPNConfig() RPNConfig {
	return RPNConfig{
		FrequencyWeight: 0.6,
		RecencyWeight:   0.4,
		RecencyDays:     90, // 3 months
	}
}

// CalculateRPN calculates the Risk Priority Number for a cluster
// RPN = (FrequencyScore * FrequencyWeight) + (RecencyScore * RecencyWeight)
func CalculateRPN(cluster *Cluster, config RPNConfig) float64 {
	if len(cluster.Problems) == 0 {
		return 0
	}

	// Frequency Score: logarithmic scaling to prevent large clusters from dominating
	frequencyScore := math.Log10(float64(len(cluster.Problems))+1) * 10

	// Recency Score: average recency of problems in cluster
	recencyScore := calculateRecencyScore(cluster, config.RecencyDays)

	// Combined RPN
	rpn := (frequencyScore * config.FrequencyWeight) + (recencyScore * config.RecencyWeight)

	// Scale to 0-100
	cluster.RPNScore = math.Min(rpn*10, 100)
	return cluster.RPNScore
}

// calculateRecencyScore computes how recent the problems in the cluster are
// Returns a score from 0 (old) to 10 (recent)
func calculateRecencyScore(cluster *Cluster, recencyDays int) float64 {
	if len(cluster.Problems) == 0 {
		return 0
	}

	now := time.Now()
	totalScore := 0.0
	validCount := 0

	for _, p := range cluster.Problems {
		if p.Tanggal == nil {
			continue
		}

		daysSince := now.Sub(*p.Tanggal).Hours() / 24
		if daysSince < 0 {
			daysSince = 0
		}

		// Exponential decay: more recent = higher score
		// Score = 10 * e^(-daysSince / recencyDays)
		score := 10 * math.Exp(-daysSince/float64(recencyDays))
		totalScore += score
		validCount++
	}

	if validCount == 0 {
		return 5 // Default middle score if no dates
	}

	return totalScore / float64(validCount)
}

// SortClustersByRPN sorts clusters by their RPN score (descending)
func SortClustersByRPN(clusters []Cluster) {
	// Simple bubble sort for small number of clusters
	n := len(clusters)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if clusters[j].RPNScore < clusters[j+1].RPNScore {
				clusters[j], clusters[j+1] = clusters[j+1], clusters[j]
			}
		}
	}
}
