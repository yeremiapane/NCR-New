package ai

import (
	"time"
)

// InsightType represents the type of insight
type InsightType string

const (
	InsightTypeTrend          InsightType = "TREND"
	InsightTypeProblem        InsightType = "PROBLEM"
	InsightTypeStatistic      InsightType = "STATISTIC"
	InsightTypeRecommendation InsightType = "RECOMMENDATION"
)

// InsightSeverity represents the importance level of an insight
type InsightSeverity string

const (
	SeverityInfo     InsightSeverity = "info"
	SeverityWarning  InsightSeverity = "warning"
	SeverityCritical InsightSeverity = "critical"
)

// Insight represents a single AI-generated insight
type Insight struct {
	Type        InsightType     `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Severity    InsightSeverity `json:"severity"`
	Data        interface{}     `json:"data,omitempty"` // Optional supporting data
}

// InsightsResponse contains the complete AI analysis response
type InsightsResponse struct {
	Insights    []Insight `json:"insights"`
	GeneratedAt time.Time `json:"generated_at"`
	Model       string    `json:"model"`
	ProcessTime float64   `json:"process_time_seconds"`
}

// AnalysisContext contains aggregated data for AI analysis
type AnalysisContext struct {
	// Summary stats
		TotalNCR        int64 `json:"total_ncr"`
	RunningCount    int64 `json:"running_count"`
	CompletedCount  int64 `json:"completed_count"`
	TerminatedCount int64 `json:"terminated_count"`
	ApprovedCount   int64 `json:"approved_count"`
	RejectedCount   int64 `json:"rejected_count"`
	TOCount         int64 `json:"to_count"`     // Material loss
	NonTOCount      int64 `json:"non_to_count"` // Rework/time loss

	// Trend data
	TrendData []TrendPoint `json:"trend_data"`

	// Top issues
	TopCategories  []CountItem   `json:"top_categories"`
	TopBrands      []CountItem   `json:"top_brands"`
	TopDepartments []CountItem   `json:"top_departments"`
	TopProblems    []ProblemItem `json:"top_problems"`

	// Filter context
	DateRange string `json:"date_range,omitempty"`
	Filters   string `json:"filters,omitempty"`
}

// TrendPoint represents a data point in trend analysis
type TrendPoint struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

// CountItem represents an item with a count
type CountItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// ProblemItem represents a problem with its details
type ProblemItem struct {
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	RPN         float64 `json:"rpn,omitempty"` // Risk Priority Number if available
}

// OllamaRequest represents the request body for Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}
