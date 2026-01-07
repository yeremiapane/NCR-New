package approval

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// splitAndTrim splits a string by comma, semicolon, or slash and trims whitespace
func splitAndTrim(s string) []string {
	// Replace semicolons and slashes with commas for unified splitting
	s = strings.ReplaceAll(s, ";", ",")
	s = strings.ReplaceAll(s, "/", ",")

	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// extractProductLine extracts the brand/product line from a full product name
// e.g., "Astral AP01 Top Hung Window" -> "Astral"
// e.g., "Crown 3T Sliding Door" -> "Crown"
// e.g., "DOOR - Solid Panel" -> "DOOR"
func extractProductLine(productName string) string {
	productName = strings.TrimSpace(productName)
	if productName == "" {
		return ""
	}

	// Known product lines/brands to match (case-insensitive)
	knownBrands := []string{
		"Astral", "Crown", "Royal", "Premium", "Standard", "Elite",
		"Classic", "Modern", "Aluminium", "Stainless", "Steel",
		"Glass", "Wood", "PVC", "UPVC", "Kayu", "Kaca",
	}

	// Check if product name starts with a known brand
	upperName := strings.ToUpper(productName)
	for _, brand := range knownBrands {
		if strings.HasPrefix(upperName, strings.ToUpper(brand)) {
			return brand
		}
	}

	// If no known brand, take the first word as the product line
	// Split by space, hyphen, or underscore
	words := strings.FieldsFunc(productName, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_'
	})

	if len(words) > 0 {
		firstWord := strings.TrimSpace(words[0])
		// Skip very short words or numbers
		if len(firstWord) >= 2 && !isNumeric(firstWord) {
			return firstWord
		}
		// Try second word if first is too short
		if len(words) > 1 {
			secondWord := strings.TrimSpace(words[1])
			if len(secondWord) >= 2 && !isNumeric(secondWord) {
				return secondWord
			}
		}
	}

	return productName
}

// isNumeric checks if a string is purely numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// Brand code to brand name mapping
// Based on FPPP format: XXX/FPPP/CODE/MM/YYYY where CODE is the brand identifier
var brandCodeMapping = map[string]string{
	"ASTA": "ASTA",
	"AST":  "ASTRAL",
	"ABO":  "ASTRAL",
	"ABX":  "ASTRAL",
	"FOR":  "FORISE",
	"PKC":  "FORISE",
	"RSD":  "RSD",
	"MAX":  "ALPHAMAX",
	"APX":  "ALPHAMAX",
	"CAR":  "CARRA",
	"RAE":  "ALLURE",
	"RAS":  "ALUPLUS",
	"HRB":  "HRB",
	"POL":  "POLARISA",
	// Add more mappings as needed
}

// extractBrandFromFPPP extracts brand name from FPPP/PO number
// Format: XXX/FPPP/CODE/MM/YYYY or XXX/PP/CODE/MM/YY (with typos)
// e.g., "011/FPPP/POL/09/2025" -> "POLARISA"
// e.g., "003/pp/pkc/10/25" -> "FORISE"
// e.g., "003/PM/CAR/X/2025" -> "CARRA"
func extractBrandFromFPPP(fpppNumber string) string {
	if fpppNumber == "" {
		return ""
	}

	// Normalize: uppercase and trim
	fpppNumber = strings.ToUpper(strings.TrimSpace(fpppNumber))

	// Split by slash
	parts := strings.Split(fpppNumber, "/")

	// We expect format: NUMBER/FPPP_TYPE/BRAND_CODE/MONTH/YEAR
	// The brand code is typically at position 2 (index 2) after splitting by "/"
	// But we need to handle variations like:
	// - 011/FPPP/POL/09/2025 (standard)
	// - 003/PP/PKC/10/25 (short type, short year)
	// - 003/PM/CAR/X/2025 (different type)

	if len(parts) < 3 {
		return ""
	}

	// Position 2 should be the brand code (after NUMBER and FPPP/PP/PM)
	brandCode := strings.TrimSpace(parts[2])

	// Skip if it's a number (might be date)
	if isNumeric(brandCode) {
		return ""
	}

	// Skip if it looks like a month (roman numerals or month numbers)
	if len(brandCode) <= 2 || brandCode == "I" || brandCode == "II" || brandCode == "III" ||
		brandCode == "IV" || brandCode == "V" || brandCode == "VI" || brandCode == "VII" ||
		brandCode == "VIII" || brandCode == "IX" || brandCode == "X" || brandCode == "XI" || brandCode == "XII" {
		return ""
	}

	// Look up the brand code in our mapping
	if brandName, ok := brandCodeMapping[brandCode]; ok {
		return brandName
	}

	// If not in mapping, return the code itself (uppercase)
	return brandCode
}

// Repository handles database operations for NCR approvals
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// UpsertApproval creates or updates an NCR approval
func (r *Repository) UpsertApproval(ctx context.Context, approval *NCRApproval) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "process_instance_id"}},
		UpdateAll: true,
	}).Create(approval).Error
}

// DeleteAttachments deletes all attachments for an approval
func (r *Repository) DeleteAttachments(ctx context.Context, approvalID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("ncr_approval_id = ?", approvalID).Delete(&NCRAttachment{}).Error
}

// CreateAttachments creates attachments in batch
func (r *Repository) CreateAttachments(ctx context.Context, attachments []NCRAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&attachments).Error
}

// GetByProcessInstanceID finds an approval by process instance ID
func (r *Repository) GetByProcessInstanceID(ctx context.Context, processInstanceID string) (*NCRApproval, error) {
	var approval NCRApproval
	err := r.db.WithContext(ctx).Where("process_instance_id = ?", processInstanceID).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// HasAnyData checks if there is any approval data in the database
func (r *Repository) HasAnyData(ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&NCRApproval{}).Limit(1).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListParams contains parameters for listing approvals
type ListParams struct {
	Page            int
	PageSize        int
	Status          string
	Search          string
	BusinessID      string
	Department      string
	DitujukanKepada string
	DilaporkanOleh  string
	Kategori        string
	ToTidakTo       string
	StartDate       *time.Time
	EndDate         *time.Time
}

// ListApprovals lists NCR approvals with filters
func (r *Repository) ListApprovals(ctx context.Context, params ListParams) ([]NCRApproval, int64, error) {
	var approvals []NCRApproval
	var total int64

	query := r.db.WithContext(ctx).Model(&NCRApproval{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.BusinessID != "" {
		query = query.Where("business_id ILIKE ?", "%"+params.BusinessID+"%")
	}
	if params.Department != "" {
		query = query.Where("originator_dept_name ILIKE ?", "%"+params.Department+"%")
	}
	if params.DitujukanKepada != "" {
		query = query.Where("ditujukan_kepada ILIKE ?", "%"+params.DitujukanKepada+"%")
	}
	if params.DilaporkanOleh != "" {
		query = query.Where("dilaporkan_oleh ILIKE ?", "%"+params.DilaporkanOleh+"%")
	}
	if params.Kategori != "" {
		query = query.Where("kategori ILIKE ?", "%"+params.Kategori+"%")
	}
	if params.ToTidakTo != "" {
		query = query.Where("to_tidak_to = ?", params.ToTidakTo)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(
			"title ILIKE ? OR "+
				"originator_name ILIKE ? OR "+
				"nama_project ILIKE ? OR "+
				"nomor_fppp ILIKE ? OR "+
				"business_id ILIKE ? OR "+
				"deskripsi_masalah ILIKE ? OR "+
				"ditujukan_kepada ILIKE ? OR "+
				"dilaporkan_oleh ILIKE ? OR "+
				"kategori ILIKE ? OR "+
				"nama_item_product ILIKE ? OR "+
				"nomor_production_order ILIKE ? OR "+
				"catatan_tambahan ILIKE ? OR "+
				"analisis_penyebab_masalah ILIKE ? OR "+
				"tindakan_perbaikan ILIKE ? OR "+
				"tindakan_pencegahan ILIKE ? OR "+
				"remark_comment ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			searchTerm,
		)
	}
	if params.StartDate != nil {
		query = query.Where("tanggal >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("tanggal <= ?", params.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("tanggal DESC, dingtalk_create_time DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&approvals).Error; err != nil {
		return nil, 0, err
	}

	return approvals, total, nil
}

// FilterOptions contains distinct values for filter dropdowns
type FilterOptions struct {
	Departments     []string `json:"departments"`
	DitujukanKepada []string `json:"ditujukan_kepada"`
	DilaporkanOleh  []string `json:"dilaporkan_oleh"`
	Kategori        []string `json:"kategori"`
	Statuses        []string `json:"statuses"`
}

// GetFilterOptions gets distinct values for filter dropdowns
func (r *Repository) GetFilterOptions(ctx context.Context) (*FilterOptions, error) {
	options := &FilterOptions{}

	// Get distinct departments
	var departments []string
	r.db.WithContext(ctx).Model(&NCRApproval{}).
		Distinct("originator_dept_name").
		Where("originator_dept_name IS NOT NULL AND originator_dept_name != ''").
		Order("originator_dept_name").
		Pluck("originator_dept_name", &departments)
	options.Departments = departments

	// Get distinct ditujukan_kepada
	var ditujukanKepada []string
	r.db.WithContext(ctx).Model(&NCRApproval{}).
		Distinct("ditujukan_kepada").
		Where("ditujukan_kepada IS NOT NULL AND ditujukan_kepada != ''").
		Order("ditujukan_kepada").
		Pluck("ditujukan_kepada", &ditujukanKepada)
	options.DitujukanKepada = ditujukanKepada

	// Get distinct dilaporkan_oleh
	var dilaporkanOleh []string
	r.db.WithContext(ctx).Model(&NCRApproval{}).
		Distinct("dilaporkan_oleh").
		Where("dilaporkan_oleh IS NOT NULL AND dilaporkan_oleh != ''").
		Order("dilaporkan_oleh").
		Pluck("dilaporkan_oleh", &dilaporkanOleh)
	options.DilaporkanOleh = dilaporkanOleh

	// Get distinct kategori
	var kategori []string
	r.db.WithContext(ctx).Model(&NCRApproval{}).
		Distinct("kategori").
		Where("kategori IS NOT NULL AND kategori != ''").
		Order("kategori").
		Pluck("kategori", &kategori)
	options.Kategori = kategori

	// Get distinct statuses
	var statuses []string
	r.db.WithContext(ctx).Model(&NCRApproval{}).
		Distinct("status").
		Where("status IS NOT NULL AND status != ''").
		Order("status").
		Pluck("status", &statuses)
	options.Statuses = statuses

	return options, nil
}

// GetApprovalWithDetails gets an approval with all related data
func (r *Repository) GetApprovalWithDetails(ctx context.Context, id uuid.UUID) (*NCRApproval, error) {
	var approval NCRApproval
	err := r.db.WithContext(ctx).
		Preload("Attachments").
		Where("id = ?", id).
		First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// StatsParams contains parameters for filtering statistics
type StatsParams struct {
	Status          string
	Search          string
	Department      string
	DitujukanKepada string
	DilaporkanOleh  string
	Kategori        string
	StartDate       *time.Time
	EndDate         *time.Time
}

// GetStatsWithFilters retrieves dashboard statistics with optional filters
func (r *Repository) GetStatsWithFilters(ctx context.Context, params StatsParams) (map[string]interface{}, error) {
	// Helper function to apply common filters
	applyFilters := func(query *gorm.DB) *gorm.DB {
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
		if params.Search != "" {
			query = query.Where("title ILIKE ? OR originator_name ILIKE ? OR nama_project ILIKE ? OR nomor_fppp ILIKE ? OR business_id ILIKE ?",
				"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
		}
		if params.Department != "" {
			query = query.Where("originator_dept_name ILIKE ?", "%"+params.Department+"%")
		}
		if params.DitujukanKepada != "" {
			query = query.Where("ditujukan_kepada ILIKE ?", "%"+params.DitujukanKepada+"%")
		}
		if params.DilaporkanOleh != "" {
			query = query.Where("dilaporkan_oleh ILIKE ?", "%"+params.DilaporkanOleh+"%")
		}
		if params.Kategori != "" {
			query = query.Where("kategori ILIKE ?", "%"+params.Kategori+"%")
		}
		if params.StartDate != nil {
			query = query.Where("tanggal >= ?", params.StartDate)
		}
		if params.EndDate != nil {
			query = query.Where("tanggal <= ?", params.EndDate)
		}
		return query
	}

	var totalCount int64
	var runningCount int64
	var completedCount int64
	var agreeCount int64
	var refuseCount int64
	var terminatedCount int64
	var toCount int64
	var tidakToCount int64

	// Count all records
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{})).Count(&totalCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("status = ?", "RUNNING")).Count(&runningCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("status = ?", "COMPLETED")).Count(&completedCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("status = ?", "TERMINATED")).Count(&terminatedCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("result = ?", "agree")).Count(&agreeCount)
	// Refuse count: check both result='refuse' AND status='TERMINATED' (terminated means rejected)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("result = ? OR status = ?", "refuse", "TERMINATED")).Count(&refuseCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("to_tidak_to ILIKE ?", "%TO%").Where("to_tidak_to NOT ILIKE ?", "%TIDAK%")).Count(&toCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("to_tidak_to ILIKE ?", "%TIDAK TO%")).Count(&tidakToCount)

	// Helper to exclude Terminated status from chart queries
	excludeTerminated := func(query *gorm.DB) *gorm.DB {
		return query.Where("status != ?", "TERMINATED")
	}

	// Get dilaporkan oleh distribution (using this instead of department)
	type DeptCount struct {
		Department string `json:"department"`
		Count      int64  `json:"count"`
	}
	var rawDilaporkanCounts []struct {
		DilaporkanOleh string
		Count          int64
	}
	dilaporkanQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("dilaporkan_oleh, count(*) as count").
		Where("dilaporkan_oleh IS NOT NULL AND dilaporkan_oleh != ''")
	excludeTerminated(applyFilters(dilaporkanQuery)).
		Group("dilaporkan_oleh").
		Order("count DESC").
		Scan(&rawDilaporkanCounts)

	// Normalize comma-separated values
	normalizedDilaporkan := make(map[string]int64)
	for _, dc := range rawDilaporkanCounts {
		parts := splitAndTrim(dc.DilaporkanOleh)
		for _, part := range parts {
			if part != "" {
				normalizedDilaporkan[part] += dc.Count
			}
		}
	}

	// Convert to slice and sort
	var deptCounts []DeptCount
	for name, count := range normalizedDilaporkan {
		deptCounts = append(deptCounts, DeptCount{
			Department: name,
			Count:      count,
		})
	}
	for i := 0; i < len(deptCounts)-1; i++ {
		for j := i + 1; j < len(deptCounts); j++ {
			if deptCounts[j].Count > deptCounts[i].Count {
				deptCounts[i], deptCounts[j] = deptCounts[j], deptCounts[i]
			}
		}
	}
	if len(deptCounts) > 10 {
		deptCounts = deptCounts[:10]
	}

	// Get category distribution (with normalization for comma-separated values)
	type KategoriCount struct {
		Kategori string `json:"kategori"`
		Count    int64  `json:"count"`
	}
	var rawKategoriCounts []KategoriCount
	kategoriQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("kategori, count(*) as count").
		Where("kategori IS NOT NULL AND kategori != ''")
	excludeTerminated(applyFilters(kategoriQuery)).
		Group("kategori").
		Order("count DESC").
		Scan(&rawKategoriCounts)

	// Normalize comma-separated values (same as ditujukan_kepada)
	normalizedKategori := make(map[string]int64)
	for _, kc := range rawKategoriCounts {
		parts := splitAndTrim(kc.Kategori)
		for _, part := range parts {
			if part != "" {
				normalizedKategori[part] += kc.Count
			}
		}
	}

	// Convert back to slice and sort by count
	var kategoriCounts []KategoriCount
	for name, count := range normalizedKategori {
		kategoriCounts = append(kategoriCounts, KategoriCount{
			Kategori: name,
			Count:    count,
		})
	}
	// Sort by count descending
	for i := 0; i < len(kategoriCounts)-1; i++ {
		for j := i + 1; j < len(kategoriCounts); j++ {
			if kategoriCounts[j].Count > kategoriCounts[i].Count {
				kategoriCounts[i], kategoriCounts[j] = kategoriCounts[j], kategoriCounts[i]
			}
		}
	}
	// Limit to top 10
	if len(kategoriCounts) > 10 {
		kategoriCounts = kategoriCounts[:10]
	}

	// Get trend data - daily or monthly counts based on filter range
	type TrendData struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}
	var trendData []TrendData
	trendQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).Where("tanggal IS NOT NULL")

	// If date range is short (<=31 days), group by day; otherwise by month
	if params.StartDate != nil && params.EndDate != nil {
		daysDiff := int(params.EndDate.Sub(*params.StartDate).Hours() / 24)
		if daysDiff <= 31 {
			// Group by day
			excludeTerminated(applyFilters(trendQuery)).
				Select("TO_CHAR(tanggal, 'YYYY-MM-DD') as month, count(*) as count").
				Group("TO_CHAR(tanggal, 'YYYY-MM-DD')").
				Order("month ASC").
				Scan(&trendData)
		} else {
			// Group by month
			excludeTerminated(applyFilters(trendQuery)).
				Select("TO_CHAR(tanggal, 'YYYY-MM') as month, count(*) as count").
				Group("TO_CHAR(tanggal, 'YYYY-MM')").
				Order("month ASC").
				Scan(&trendData)
		}
	} else {
		// Default: group by month
		excludeTerminated(applyFilters(trendQuery)).
			Select("TO_CHAR(tanggal, 'YYYY-MM') as month, count(*) as count").
			Group("TO_CHAR(tanggal, 'YYYY-MM')").
			Order("month ASC").
			Scan(&trendData)
	}

	// Get ditujukan kepada distribution (with normalization for comma-separated values)
	type DitujukanCount struct {
		DitujukanKepada string `json:"ditujukan_kepada"`
		Count           int64  `json:"count"`
	}
	var rawDitujukanCounts []DitujukanCount
	ditujukanQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("ditujukan_kepada, count(*) as count").
		Where("ditujukan_kepada IS NOT NULL AND ditujukan_kepada != ''")
	excludeTerminated(applyFilters(ditujukanQuery)).
		Group("ditujukan_kepada").
		Order("count DESC").
		Scan(&rawDitujukanCounts)

	// Normalize comma-separated values (e.g., "Klaes,RND" -> ["Klaes", "RND"])
	normalizedCounts := make(map[string]int64)
	for _, dc := range rawDitujukanCounts {
		// Split by comma, semicolon, or slash
		parts := splitAndTrim(dc.DitujukanKepada)
		for _, part := range parts {
			if part != "" {
				normalizedCounts[part] += dc.Count
			}
		}
	}

	// Convert back to slice and sort by count
	var ditujukanCounts []DitujukanCount
	for name, count := range normalizedCounts {
		ditujukanCounts = append(ditujukanCounts, DitujukanCount{
			DitujukanKepada: name,
			Count:           count,
		})
	}
	// Sort by count descending
	for i := 0; i < len(ditujukanCounts)-1; i++ {
		for j := i + 1; j < len(ditujukanCounts); j++ {
			if ditujukanCounts[j].Count > ditujukanCounts[i].Count {
				ditujukanCounts[i], ditujukanCounts[j] = ditujukanCounts[j], ditujukanCounts[i]
			}
		}
	}
	// Limit to top 10
	if len(ditujukanCounts) > 10 {
		ditujukanCounts = ditujukanCounts[:10]
	}

	// Get brand distribution from FPPP number (with fallback to production order)
	// Extract brand code from format like: 011/FPPP/POL/09/2025 -> POLARISA
	type BrandCount struct {
		NomorFPPP            string `gorm:"column:nomor_fppp"`
		NomorProductionOrder string `gorm:"column:nomor_production_order"`
		Count                int64  `json:"count"`
	}
	var rawBrandCounts []BrandCount
	brandQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("nomor_fppp, nomor_production_order, count(*) as count").
		Where("(nomor_fppp IS NOT NULL AND nomor_fppp != '') OR (nomor_production_order IS NOT NULL AND nomor_production_order != '')")
	excludeTerminated(applyFilters(brandQuery)).
		Group("nomor_fppp, nomor_production_order").
		Order("count DESC").
		Scan(&rawBrandCounts)

	// Normalize by extracting brand from FPPP/PO number
	type ItemProductCount struct {
		NamaItemProduct string `json:"nama_item_product"`
		Count           int64  `json:"count"`
	}
	normalizedItemProduct := make(map[string]int64)
	for _, bc := range rawBrandCounts {
		// Try FPPP first, fallback to production order
		fpppNumber := bc.NomorFPPP
		if fpppNumber == "" {
			fpppNumber = bc.NomorProductionOrder
		}

		if fpppNumber != "" {
			brand := extractBrandFromFPPP(fpppNumber)
			if brand != "" {
				normalizedItemProduct[brand] += bc.Count
			}
		}
	}

	// Convert back to slice and sort by count
	var itemProductCounts []ItemProductCount
	for name, count := range normalizedItemProduct {
		itemProductCounts = append(itemProductCounts, ItemProductCount{
			NamaItemProduct: name,
			Count:           count,
		})
	}
	// Sort by count descending
	for i := 0; i < len(itemProductCounts)-1; i++ {
		for j := i + 1; j < len(itemProductCounts); j++ {
			if itemProductCounts[j].Count > itemProductCounts[i].Count {
				itemProductCounts[i], itemProductCounts[j] = itemProductCounts[j], itemProductCounts[i]
			}
		}
	}
	// Limit to top 10
	if len(itemProductCounts) > 10 {
		itemProductCounts = itemProductCounts[:10]
	}

	return map[string]interface{}{
		"total":                    totalCount,
		"running":                  runningCount,
		"completed":                completedCount,
		"terminated":               terminatedCount,
		"approved":                 agreeCount,
		"rejected":                 refuseCount,
		"to":                       toCount,
		"tidak_to":                 tidakToCount,
		"department_counts":        deptCounts,
		"kategori_counts":          kategoriCounts,
		"ditujukan_kepada_counts":  ditujukanCounts,
		"nama_item_product_counts": itemProductCounts,
		"trend_data":               trendData,
	}, nil
}

// CreateSyncLog creates a sync log entry
func (r *Repository) CreateSyncLog(ctx context.Context, log *SyncLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateSyncLog updates a sync log entry
func (r *Repository) UpdateSyncLog(ctx context.Context, log *SyncLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// ListSyncLogs lists sync logs with pagination
func (r *Repository) ListSyncLogs(ctx context.Context, page, pageSize int) ([]SyncLog, int64, error) {
	var logs []SyncLog
	var total int64

	r.db.WithContext(ctx).Model(&SyncLog{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("started_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}
