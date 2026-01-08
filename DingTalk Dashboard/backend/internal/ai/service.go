package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dingtalk-dashboard/internal/domain/approval"

	"go.uber.org/zap"
)

// Service orchestrates AI insights generation
type Service struct {
	ollamaClient *OllamaClient
	approvalRepo *approval.Repository
	logger       *zap.Logger
}

// NewService creates a new AI service
func NewService(ollamaClient *OllamaClient, approvalRepo *approval.Repository, logger *zap.Logger) *Service {
	return &Service{
		ollamaClient: ollamaClient,
		approvalRepo: approvalRepo,
		logger:       logger,
	}
}

// GenerateInsights generates AI insights based on current dashboard data
func (s *Service) GenerateInsights(ctx context.Context, params approval.StatsParams) (*InsightsResponse, error) {
	startTime := time.Now()

	// Check Ollama health first
	if err := s.ollamaClient.CheckHealth(ctx); err != nil {
		return nil, fmt.Errorf("Ollama service not available: %w", err)
	}

	// Get dashboard statistics
	stats, err := s.approvalRepo.GetStatsWithFilters(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Get recent problems with descriptions for context
	recentProblems, err := s.getRecentProblems(ctx, params)
	if err != nil {
		s.logger.Warn("Failed to get recent problems", zap.Error(err))
		// Continue without problem details
	}

	// Build analysis context from stats
	analysisCtx := s.buildAnalysisContext(stats, params)
	analysisCtx.TopProblems = recentProblems

	// Generate prompt
	userPrompt := BuildAnalysisPrompt(analysisCtx)

	s.logger.Info("Generating AI insights",
		zap.String("model", s.ollamaClient.GetModel()),
		zap.Int64("total_ncr", analysisCtx.TotalNCR),
		zap.Int("problem_samples", len(recentProblems)),
	)

	// Call Ollama
	response, err := s.ollamaClient.Generate(ctx, SystemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate insights: %w", err)
	}

	// Parse insights from response
	insights, err := s.parseInsights(response)
	if err != nil {
		s.logger.Warn("Failed to parse AI response, using raw response",
			zap.Error(err),
			zap.String("raw_response", response),
		)
		// Fallback: create a single insight with the raw response
		insights = []Insight{{
			Type:        InsightTypeStatistic,
			Title:       "AI Analysis",
			Description: response,
			Severity:    SeverityInfo,
		}}
	}

	processTime := time.Since(startTime).Seconds()

	return &InsightsResponse{
		Insights:    insights,
		GeneratedAt: time.Now(),
		Model:       s.ollamaClient.GetModel(),
		ProcessTime: processTime,
	}, nil
}

// getRecentProblems fetches recent NCR problems with their descriptions for AI context
func (s *Service) getRecentProblems(ctx context.Context, params approval.StatsParams) ([]ProblemItem, error) {
	// Convert StatsParams to ListParams for fetching approvals
	listParams := approval.ListParams{
		Page:            1,
		PageSize:        20, // Get top 20 recent problems for context
		Status:          params.Status,
		Search:          params.Search,
		Department:      params.Department,
		DitujukanKepada: params.DitujukanKepada,
		DilaporkanOleh:  params.DilaporkanOleh,
		Kategori:        params.Kategori,
		StartDate:       params.StartDate,
		EndDate:         params.EndDate,
	}

	approvals, _, err := s.approvalRepo.ListApprovals(ctx, listParams)
	if err != nil {
		return nil, err
	}

	var problems []ProblemItem
	for _, a := range approvals {
		// Skip if no description
		if a.DeskripsiMasalah == "" {
			continue
		}

		// Combine description, analysis, and remarks for full context
		description := a.DeskripsiMasalah
		if a.AnalisisPenyebabMasalah != "" {
			description += " | Analysis: " + a.AnalisisPenyebabMasalah
		}
		if a.RemarkComment != "" && len(a.RemarkComment) < 200 {
			description += " | Remark: " + a.RemarkComment
		}

		// Truncate if too long
		if len(description) > 300 {
			description = description[:297] + "..."
		}

		problems = append(problems, ProblemItem{
			Description: description,
			Category:    a.Kategori,
			Brand:       a.NamaItemProduct,
		})

		// Limit to 10 for prompt size
		if len(problems) >= 10 {
			break
		}
	}

	return problems, nil
}

// buildAnalysisContext converts stats response to AnalysisContext
func (s *Service) buildAnalysisContext(stats map[string]interface{}, params approval.StatsParams) AnalysisContext {
	ctx := AnalysisContext{}

	// Extract summary counts
	if v, ok := stats["total"].(int64); ok {
		ctx.TotalNCR = v
	}
	if v, ok := stats["running"].(int64); ok {
		ctx.RunningCount = v
	}
	if v, ok := stats["completed"].(int64); ok {
		ctx.CompletedCount = v
	}
	if v, ok := stats["terminated"].(int64); ok {
		ctx.TerminatedCount = v
	}
	if v, ok := stats["approved"].(int64); ok {
		ctx.ApprovedCount = v
	}
	if v, ok := stats["rejected"].(int64); ok {
		ctx.RejectedCount = v
	}
	if v, ok := stats["to"].(int64); ok {
		ctx.TOCount = v
	}
	if v, ok := stats["tidak_to"].(int64); ok {
		ctx.NonTOCount = v
	}

	// Extract trend data
	if trendData, ok := stats["trend_data"].([]struct {
		Month string
		Count int64
	}); ok {
		for _, t := range trendData {
			ctx.TrendData = append(ctx.TrendData, TrendPoint{
				Period: t.Month,
				Count:  t.Count,
			})
		}
	}

	// Extract top categories using type assertion for the actual struct type
	if kategoriCounts, ok := stats["kategori_counts"]; ok {
		ctx.TopCategories = s.extractCountItems(kategoriCounts, "kategori")
	}

	// Extract top brands
	if itemProductCounts, ok := stats["nama_item_product_counts"]; ok {
		ctx.TopBrands = s.extractCountItems(itemProductCounts, "nama_item_product")
	}

	// Extract top departments
	if deptCounts, ok := stats["department_counts"]; ok {
		ctx.TopDepartments = s.extractCountItems(deptCounts, "department")
	}

	// Build date range string
	if params.StartDate != nil && params.EndDate != nil {
		ctx.DateRange = fmt.Sprintf("%s to %s",
			params.StartDate.Format("2006-01-02"),
			params.EndDate.Format("2006-01-02"),
		)
	}

	// Build filter context
	var filters []string
	if params.Department != "" {
		filters = append(filters, "Department: "+params.Department)
	}
	if params.Kategori != "" {
		filters = append(filters, "Category: "+params.Kategori)
	}
	if params.DitujukanKepada != "" {
		filters = append(filters, "Assigned To: "+params.DitujukanKepada)
	}
	if len(filters) > 0 {
		ctx.Filters = strings.Join(filters, ", ")
	}

	return ctx
}

// extractCountItems extracts count items from interface{} using JSON marshaling
func (s *Service) extractCountItems(data interface{}, fieldType string) []CountItem {
	var items []CountItem

	// Marshal to JSON and unmarshal to a generic structure
	jsonData, err := json.Marshal(data)
	if err != nil {
		return items
	}

	var rawItems []map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawItems); err != nil {
		return items
	}

	for _, item := range rawItems {
		var name string
		var count int64

		// Try different field names based on type
		switch fieldType {
		case "kategori":
			if v, ok := item["kategori"].(string); ok {
				name = v
			}
		case "nama_item_product":
			if v, ok := item["nama_item_product"].(string); ok {
				name = v
			}
		case "department":
			if v, ok := item["department"].(string); ok {
				name = v
			}
		}

		if v, ok := item["count"].(float64); ok {
			count = int64(v)
		}

		if name != "" && count > 0 {
			items = append(items, CountItem{Name: name, Count: count})
		}
	}

	return items
}

// parseInsights parses the LLM response into structured insights
func (s *Service) parseInsights(response string) ([]Insight, error) {
	// Clean the response - remove potential markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find the JSON array boundaries
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil, fmt.Errorf("no valid JSON array found in response")
	}

	jsonStr := response[startIdx : endIdx+1]

	var insights []Insight
	if err := json.Unmarshal([]byte(jsonStr), &insights); err != nil {
		return nil, fmt.Errorf("failed to parse insights JSON: %w", err)
	}

	// Validate and normalize insights
	validInsights := make([]Insight, 0, len(insights))
	for _, insight := range insights {
		// Validate type
		switch insight.Type {
		case InsightTypeTrend, InsightTypeProblem, InsightTypeStatistic, InsightTypeRecommendation:
			// Valid
		default:
			insight.Type = InsightTypeStatistic // Default to statistic
		}

		// Validate severity
		switch insight.Severity {
		case SeverityInfo, SeverityWarning, SeverityCritical:
			// Valid
		default:
			insight.Severity = SeverityInfo // Default to info
		}

		if insight.Title != "" && insight.Description != "" {
			validInsights = append(validInsights, insight)
		}
	}

	if len(validInsights) == 0 {
		return nil, fmt.Errorf("no valid insights parsed from response")
	}

	return validInsights, nil
}

// CheckHealth checks if the AI service is available
func (s *Service) CheckHealth(ctx context.Context) error {
	return s.ollamaClient.CheckHealth(ctx)
}
