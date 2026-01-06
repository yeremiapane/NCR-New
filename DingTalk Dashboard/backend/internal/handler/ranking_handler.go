package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"dingtalk-dashboard/internal/ranking"
)

// RankingHandler handles problem ranking endpoints
type RankingHandler struct {
	service *ranking.Service
}

// NewRankingHandler creates a new ranking handler
func NewRankingHandler(service *ranking.Service) *RankingHandler {
	return &RankingHandler{service: service}
}

// parseRankingFilters extracts filter params from request
func parseRankingFilters(c *fiber.Ctx) ranking.RankingFilters {
	filters := ranking.RankingFilters{
		Department:      c.Query("department"),
		Kategori:        c.Query("kategori"),
		DitujukanKepada: c.Query("ditujukan_kepada"),
		DilaporkanOleh:  c.Query("dilaporkan_oleh"),
		Status:          c.Query("status"),
		Search:          c.Query("search"),
	}

	// Parse date filters
	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			filters.StartDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			filters.EndDate = &t
		}
	}

	return filters
}

// GetProblemRanking handles GET /api/v1/approvals/problem-ranking
func (h *RankingHandler) GetProblemRanking(c *fiber.Ctx) error {
	filters := parseRankingFilters(c)

	// Check if debug mode is requested
	debug := c.Query("debug") == "true"

	// Get top 6 problems with optional stats
	problems, stats, err := h.service.GetTopProblemsWithStats(c.Context(), 6, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get problem ranking",
			"error":   err.Error(),
		})
	}

	// Log stats to console
	if stats != nil {
		fmt.Printf("[RANKING] Problems: %d, Clusters: %d, Vocabulary: %d, Top terms: %v\n",
			stats.TotalProblems, stats.ClusterCount, stats.VocabularySize, stats.TopTerms)
	}

	response := fiber.Map{
		"success": true,
		"message": "Problem ranking fetched successfully",
		"data":    problems,
	}

	// Include stats in response if debug mode
	if debug && stats != nil {
		response["algorithm_stats"] = stats
	}

	return c.JSON(response)
}

// GetWordCloud handles GET /api/v1/approvals/word-cloud
func (h *RankingHandler) GetWordCloud(c *fiber.Ctx) error {
	filters := parseRankingFilters(c)

	// Get word frequencies for word cloud (top 30 words)
	wordFreqs, err := h.service.GetWordCloud(c.Context(), 30, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get word cloud data",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Word cloud data fetched successfully",
		"data":    wordFreqs,
	})
}

// GetRankingDebug handles GET /api/v1/approvals/ranking-debug
// Returns detailed similarity scores between problems
func (h *RankingHandler) GetRankingDebug(c *fiber.Ctx) error {
	filters := parseRankingFilters(c)

	// Get debug info
	debugInfo, err := h.service.GetRankingDebugInfo(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get ranking debug info",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Ranking debug info fetched successfully",
		"data":    debugInfo,
	})
}
