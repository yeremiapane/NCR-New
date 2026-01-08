package handler

import (
	"time"

	"dingtalk-dashboard/internal/ai"
	"dingtalk-dashboard/internal/domain/approval"

	"github.com/gofiber/fiber/v2"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	aiService *ai.Service
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService *ai.Service) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// GetInsights handles GET /api/v1/ai/insights
func (h *AIHandler) GetInsights(c *fiber.Ctx) error {
	// Parse filter parameters (same as stats endpoint)
	params := approval.StatsParams{
		Status:          c.Query("status"),
		Search:          c.Query("search"),
		Department:      c.Query("department"),
		DitujukanKepada: c.Query("ditujukan_kepada"),
		DilaporkanOleh:  c.Query("dilaporkan_oleh"),
		Kategori:        c.Query("kategori"),
	}

	// Parse date filters
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			t = t.Add(24*time.Hour - time.Second) // End of day
			params.EndDate = &t
		}
	}

	// Generate insights
	insights, err := h.aiService.GenerateInsights(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate AI insights",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "AI insights generated successfully",
		"data":    insights,
	})
}

// CheckHealth handles GET /api/v1/ai/health
func (h *AIHandler) CheckHealth(c *fiber.Ctx) error {
	if err := h.aiService.CheckHealth(c.Context()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "AI service not available",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "AI service is healthy",
	})
}
