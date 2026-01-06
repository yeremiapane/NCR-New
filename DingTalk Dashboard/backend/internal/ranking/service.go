package ranking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service provides problem ranking functionality
type Service struct {
	db        *gorm.DB
	rpnConfig RPNConfig
	threshold float64 // Similarity threshold for clustering
}

// NewService creates a new ranking service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:        db,
		rpnConfig: DefaultRPNConfig(),
		threshold: 0.15, // 15% similarity threshold - lower for semantic matching
	}
}

// RankingFilters contains all filter parameters
type RankingFilters struct {
	Department      string
	Kategori        string
	DitujukanKepada string
	DilaporkanOleh  string
	Status          string
	Search          string
	StartDate       *time.Time
	EndDate         *time.Time
}

// NCRApprovalForRanking represents the data needed for ranking
type NCRApprovalForRanking struct {
	ID               uuid.UUID  `gorm:"column:id"`
	DeskripsiMasalah string     `gorm:"column:deskripsi_masalah"`
	Tanggal          *time.Time `gorm:"column:tanggal"`
	Status           string     `gorm:"column:status"`
	Result           string     `gorm:"column:result"`
	Kategori         string     `gorm:"column:kategori"`
	DepartmentName   string     `gorm:"column:originator_dept_name"`
}

// TableName specifies the table name for GORM
func (NCRApprovalForRanking) TableName() string {
	return "ncr_approvals"
}

// applyFilters adds WHERE clauses based on filter params
func (s *Service) applyFilters(query *gorm.DB, filters RankingFilters) *gorm.DB {
	if filters.Department != "" {
		query = query.Where("originator_dept_name ILIKE ?", "%"+filters.Department+"%")
	}
	if filters.Kategori != "" {
		query = query.Where("kategori ILIKE ?", "%"+filters.Kategori+"%")
	}
	if filters.DitujukanKepada != "" {
		query = query.Where("ditujukan_kepada ILIKE ?", "%"+filters.DitujukanKepada+"%")
	}
	if filters.DilaporkanOleh != "" {
		query = query.Where("dilaporkan_oleh ILIKE ?", "%"+filters.DilaporkanOleh+"%")
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		query = query.Where("title ILIKE ? OR originator_name ILIKE ? OR nama_project ILIKE ? OR deskripsi_masalah ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}
	if filters.StartDate != nil {
		query = query.Where("tanggal >= ?", filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("tanggal <= ?", filters.EndDate)
	}
	return query
}

// fetchProblems fetches problems from DB with filters
func (s *Service) fetchProblems(ctx context.Context, filters RankingFilters) ([]ProblemData, error) {
	query := s.db.WithContext(ctx).
		Select("id, deskripsi_masalah, tanggal, status, result, kategori, originator_dept_name").
		Where("deskripsi_masalah IS NOT NULL AND deskripsi_masalah != ''")

	query = s.applyFilters(query, filters)

	var approvals []NCRApprovalForRanking
	err := query.Find(&approvals).Error
	if err != nil {
		return nil, err
	}

	// Convert to ProblemData
	problems := make([]ProblemData, len(approvals))
	for i, a := range approvals {
		problems[i] = ProblemData{
			ID:               a.ID,
			DeskripsiMasalah: a.DeskripsiMasalah,
			Tanggal:          a.Tanggal,
			Status:           a.Status,
			Result:           a.Result,
			Kategori:         a.Kategori,
			Department:       a.DepartmentName,
		}
	}

	return problems, nil
}

// GetTopProblems returns the top N ranked problem clusters
func (s *Service) GetTopProblems(ctx context.Context, limit int) ([]RankedProblem, error) {
	result, _, err := s.GetTopProblemsWithStats(ctx, limit, RankingFilters{})
	return result, err
}

// GetTopProblemsFiltered returns top problems with optional filters
func (s *Service) GetTopProblemsFiltered(ctx context.Context, limit int, filters RankingFilters) ([]RankedProblem, error) {
	result, _, err := s.GetTopProblemsWithStats(ctx, limit, filters)
	return result, err
}

// GetTopProblemsWithStats returns top problems with clustering stats
func (s *Service) GetTopProblemsWithStats(ctx context.Context, limit int, filters RankingFilters) ([]RankedProblem, *ClusterStats, error) {
	problems, err := s.fetchProblems(ctx, filters)
	if err != nil {
		return nil, nil, err
	}

	if len(problems) == 0 {
		return []RankedProblem{}, &ClusterStats{}, nil
	}

	// Cluster with stats
	clusters, stats := ClusterDescriptionsSemanticWithStats(problems, s.threshold)

	// Calculate RPN and select centroids
	for i := range clusters {
		SelectCentroid(&clusters[i])
		CalculateRPN(&clusters[i], s.rpnConfig)
	}

	// Sort and limit
	SortClustersByRPN(clusters)
	if len(clusters) > limit {
		clusters = clusters[:limit]
	}

	// Convert to output with simplified key phrase descriptions
	result := make([]RankedProblem, len(clusters))
	for i, c := range clusters {
		result[i] = RankedProblem{
			Rank:          i + 1,
			Description:   c.GetClusterKeyPhrase(4), // Max 4 words summary
			Frequency:     len(c.Problems),
			RPNScore:      c.RPNScore,
			Kategori:      c.GetMostCommonKategori(),
			SampleIDs:     c.GetSampleIDs(),
			AlgorithmInfo: fmt.Sprintf("Vocab: %d, Cluster size: %d", stats.VocabularySize, len(c.Problems)),
		}
	}

	return result, stats, nil
}

// GetWordCloud returns word frequencies for word cloud visualization
func (s *Service) GetWordCloud(ctx context.Context, limit int, filters RankingFilters) ([]WordFrequency, error) {
	problems, err := s.fetchProblems(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Extract all descriptions
	var texts []string
	for _, p := range problems {
		texts = append(texts, p.DeskripsiMasalah)
	}

	// Count word frequencies
	return CountWordFrequencies(texts, limit), nil
}

// DebugSimilarityPair represents similarity between two problems
type DebugSimilarityPair struct {
	Problem1ID  string  `json:"problem1_id"`
	Problem1    string  `json:"problem1"`
	Problem2ID  string  `json:"problem2_id"`
	Problem2    string  `json:"problem2"`
	TrigramSim  float64 `json:"trigram_similarity"`
	LCSSim      float64 `json:"lcs_similarity"`
	TFIDFSim    float64 `json:"tfidf_similarity"`
	CombinedSim float64 `json:"combined_similarity"`
}

// RankingDebugInfo contains detailed debug information
type RankingDebugInfo struct {
	Stats           *ClusterStats         `json:"stats"`
	SimilarityPairs []DebugSimilarityPair `json:"similarity_pairs"`
}

// GetRankingDebugInfo returns detailed debug info about similarity calculations
func (s *Service) GetRankingDebugInfo(ctx context.Context, filters RankingFilters) (*RankingDebugInfo, error) {
	problems, err := s.fetchProblems(ctx, filters)
	if err != nil {
		return nil, err
	}

	if len(problems) == 0 {
		return &RankingDebugInfo{Stats: &ClusterStats{}}, nil
	}

	// Get stats
	_, stats := ClusterDescriptionsSemanticWithStats(problems, s.threshold)

	// Build semantic similarity
	descriptions := make([]string, len(problems))
	for i, p := range problems {
		descriptions[i] = p.DeskripsiMasalah
	}
	semSim := NewSemanticSimilarity(descriptions)

	// Pre-compute trigrams
	for i := range problems {
		if problems[i].Trigrams == nil {
			problems[i].Trigrams = GenerateTrigrams(problems[i].DeskripsiMasalah)
		}
		problems[i].TFIDFVector = semSim.tfidfVectors[i]
	}

	// Calculate similarity pairs (limit to first 10 problems for performance)
	limit := 10
	if len(problems) < limit {
		limit = len(problems)
	}

	var pairs []DebugSimilarityPair
	for i := 0; i < limit; i++ {
		for j := i + 1; j < limit; j++ {
			trigramSim := CalculateSimilarity(problems[i].Trigrams, problems[j].Trigrams)
			lcsSim := CalculateLCSSimilarity(problems[i].DeskripsiMasalah, problems[j].DeskripsiMasalah)
			tfidfSim := CosineSimilarity(problems[i].TFIDFVector, problems[j].TFIDFVector)
			combined := (trigramSim * 0.25) + (lcsSim * 0.15) + (tfidfSim * 0.60)

			pairs = append(pairs, DebugSimilarityPair{
				Problem1ID:  problems[i].ID.String(),
				Problem1:    truncateString(problems[i].DeskripsiMasalah, 50),
				Problem2ID:  problems[j].ID.String(),
				Problem2:    truncateString(problems[j].DeskripsiMasalah, 50),
				TrigramSim:  trigramSim,
				LCSSim:      lcsSim,
				TFIDFSim:    tfidfSim,
				CombinedSim: combined,
			})
		}
	}

	return &RankingDebugInfo{
		Stats:           stats,
		SimilarityPairs: pairs,
	}, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
