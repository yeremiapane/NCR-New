package approval

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
		query = query.Where("title ILIKE ? OR originator_name ILIKE ? OR nama_project ILIKE ? OR nomor_fppp ILIKE ? OR business_id ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
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
	var toCount int64
	var tidakToCount int64

	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{})).Count(&totalCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("status = ?", "RUNNING")).Count(&runningCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("status = ?", "COMPLETED")).Count(&completedCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("result = ?", "agree")).Count(&agreeCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("result = ?", "refuse")).Count(&refuseCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("to_tidak_to ILIKE ?", "%TO%").Where("to_tidak_to NOT ILIKE ?", "%TIDAK%")).Count(&toCount)
	applyFilters(r.db.WithContext(ctx).Model(&NCRApproval{}).Where("to_tidak_to ILIKE ?", "%TIDAK TO%")).Count(&tidakToCount)

	// Get department distribution
	type DeptCount struct {
		Department string `json:"department"`
		Count      int64  `json:"count"`
	}
	var deptCounts []DeptCount
	deptQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("originator_dept_name as department, count(*) as count").
		Where("originator_dept_name IS NOT NULL AND originator_dept_name != ''")
	applyFilters(deptQuery).
		Group("originator_dept_name").
		Order("count DESC").
		Limit(10).
		Scan(&deptCounts)

	// Get category distribution
	type KategoriCount struct {
		Kategori string `json:"kategori"`
		Count    int64  `json:"count"`
	}
	var kategoriCounts []KategoriCount
	kategoriQuery := r.db.WithContext(ctx).Model(&NCRApproval{}).
		Select("kategori, count(*) as count").
		Where("kategori IS NOT NULL AND kategori != ''")
	applyFilters(kategoriQuery).
		Group("kategori").
		Order("count DESC").
		Limit(10).
		Scan(&kategoriCounts)

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
			applyFilters(trendQuery).
				Select("TO_CHAR(tanggal, 'YYYY-MM-DD') as month, count(*) as count").
				Group("TO_CHAR(tanggal, 'YYYY-MM-DD')").
				Order("month ASC").
				Scan(&trendData)
		} else {
			// Group by month
			applyFilters(trendQuery).
				Select("TO_CHAR(tanggal, 'YYYY-MM') as month, count(*) as count").
				Group("TO_CHAR(tanggal, 'YYYY-MM')").
				Order("month ASC").
				Scan(&trendData)
		}
	} else {
		// Default: group by month
		applyFilters(trendQuery).
			Select("TO_CHAR(tanggal, 'YYYY-MM') as month, count(*) as count").
			Group("TO_CHAR(tanggal, 'YYYY-MM')").
			Order("month ASC").
			Scan(&trendData)
	}

	return map[string]interface{}{
		"total":             totalCount,
		"running":           runningCount,
		"completed":         completedCount,
		"approved":          agreeCount,
		"rejected":          refuseCount,
		"to":                toCount,
		"tidak_to":          tidakToCount,
		"department_counts": deptCounts,
		"kategori_counts":   kategoriCounts,
		"trend_data":        trendData,
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
