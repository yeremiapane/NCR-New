package handler

import (
	"strconv"
	"time"

	"dingtalk-dashboard/internal/domain/approval"
	"dingtalk-dashboard/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ApprovalHandler handles HTTP requests for approvals
type ApprovalHandler struct {
	service   *approval.Service
	scheduler *scheduler.Scheduler
}

// NewApprovalHandler creates a new handler
func NewApprovalHandler(service *approval.Service, scheduler *scheduler.Scheduler) *ApprovalHandler {
	return &ApprovalHandler{
		service:   service,
		scheduler: scheduler,
	}
}

// ListApprovals handles GET /api/v1/approvals
func (h *ApprovalHandler) ListApprovals(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := approval.ListParams{
		Page:            page,
		PageSize:        pageSize,
		Status:          c.Query("status"),
		Search:          c.Query("search"),
		BusinessID:      c.Query("business_id"),
		Department:      c.Query("department"),
		DitujukanKepada: c.Query("ditujukan_kepada"),
		DilaporkanOleh:  c.Query("dilaporkan_oleh"),
		Kategori:        c.Query("kategori"),
		ToTidakTo:       c.Query("to_tidak_to"),
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

	approvals, total, err := h.service.ListApprovals(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch approvals",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Approvals fetched successfully",
		"data": fiber.Map{
			"approvals": approvals,
			"pagination": fiber.Map{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}

// GetFilterOptions handles GET /api/v1/approvals/filter-options
func (h *ApprovalHandler) GetFilterOptions(c *fiber.Ctx) error {
	options, err := h.service.GetFilterOptions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch filter options",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Filter options fetched successfully",
		"data":    options,
	})
}

// GetApproval handles GET /api/v1/approvals/:id
func (h *ApprovalHandler) GetApproval(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid approval ID",
		})
	}

	approvalInstance, err := h.service.GetApproval(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Approval not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Approval fetched successfully",
		"data":    approvalInstance,
	})
}

// GetStats handles GET /api/v1/approvals/stats
func (h *ApprovalHandler) GetStats(c *fiber.Ctx) error {
	// Parse filter parameters
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

	stats, err := h.service.GetStatsWithFilters(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch statistics",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Statistics fetched successfully",
		"data":    stats,
	})
}

// TriggerSync handles POST /api/v1/sync/trigger
func (h *ApprovalHandler) TriggerSync(c *fiber.Ctx) error {
	syncLog, err := h.scheduler.RunManualSync(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to trigger sync",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sync triggered successfully",
		"data":    syncLog,
	})
}

// ListSyncLogs handles GET /api/v1/sync/logs
func (h *ApprovalHandler) ListSyncLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.service.ListSyncLogs(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch sync logs",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sync logs fetched successfully",
		"data": fiber.Map{
			"logs": logs,
			"pagination": fiber.Map{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}
