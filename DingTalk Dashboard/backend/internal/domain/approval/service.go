package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dingtalk-dashboard/internal/dingtalk"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles approval business logic
type Service struct {
	repo   *Repository
	client *dingtalk.Client
	logger *zap.Logger
}

// NewService creates a new approval service
func NewService(repo *Repository, client *dingtalk.Client, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		client: client,
		logger: logger,
	}
}

// SyncApprovals syncs approvals from DingTalk
func (s *Service) SyncApprovals(ctx context.Context, processCode string, syncType string) (*SyncLog, error) {
	// Create sync log
	syncLog := &SyncLog{
		ID:       uuid.New(),
		SyncType: syncType,
		Status:   "started",
	}
	if err := s.repo.CreateSyncLog(ctx, syncLog); err != nil {
		return nil, err
	}

	// Determine start time based on existing data
	var startTime time.Time

	// Check if database has any data
	hasData, err := s.repo.HasAnyData(ctx)
	if err != nil {
		s.logger.Error("Failed to check existing data", zap.Error(err))
		hasData = false
	}

	if hasData {
		// If data exists, fetch from 5 days ago
		startTime = time.Now().AddDate(0, 0, -5)
		s.logger.Info("Database has data, syncing from 5 days ago", zap.Time("start_time", startTime))
	} else {
		// If no data, fetch from November 1, 2025
		startTime = time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
		s.logger.Info("Database is empty, syncing from November 1, 2025", zap.Time("start_time", startTime))
	}

	var allInstanceIDs []string
	var cursor int64 = 0

	// Fetch all instance IDs with pagination (no endTime)
	for {
		resp, err := s.client.GetApprovalInstanceIDs(processCode, startTime, cursor, 20)
		if err != nil {
			s.logger.Error("Failed to fetch instance IDs", zap.Error(err))
			syncLog.Status = "failed"
			syncLog.ErrorMessage = err.Error()
			now := time.Now()
			syncLog.CompletedAt = &now
			s.repo.UpdateSyncLog(ctx, syncLog)
			return syncLog, err
		}

		allInstanceIDs = append(allInstanceIDs, resp.Result.List...)

		if resp.Result.NextCursor == 0 || len(resp.Result.List) == 0 {
			break
		}
		cursor = resp.Result.NextCursor
	}

	s.logger.Info("Fetched instance IDs", zap.Int("count", len(allInstanceIDs)))

	// Cache for user names
	userNameCache := make(map[string]string)
	created := 0
	updated := 0

	// Process each instance
	for _, instanceID := range allInstanceIDs {
		detail, err := s.client.GetApprovalInstanceDetail(instanceID)
		if err != nil {
			s.logger.Error("Failed to fetch instance detail",
				zap.String("instance_id", instanceID),
				zap.Error(err))
			continue
		}

		// Skip if no process instance data
		if detail.ProcessInstance == nil {
			s.logger.Warn("No process instance data",
				zap.String("instance_id", instanceID))
			continue
		}

		// Check if exists
		existing, _ := s.repo.GetByProcessInstanceID(ctx, instanceID)
		isNew := existing == nil

		// Get originator name via DingTalk User API
		originatorName := s.client.GetUserName(detail.ProcessInstance.OriginatorUserID, userNameCache)

		// Create NCR approval with mapped fields
		approval := &NCRApproval{
			ProcessInstanceID:  instanceID,
			BusinessID:         detail.ProcessInstance.BusinessID,
			Title:              detail.ProcessInstance.Title,
			Status:             detail.ProcessInstance.Status,
			Result:             detail.ProcessInstance.Result,
			OriginatorUserID:   detail.ProcessInstance.OriginatorUserID,
			OriginatorName:     originatorName,
			OriginatorDeptID:   detail.ProcessInstance.OriginatorDeptID,
			OriginatorDeptName: detail.ProcessInstance.OriginatorDeptName,
			DingTalkCreateTime: dingtalk.ParseDingTalkTime(detail.ProcessInstance.CreateTime),
			DingTalkFinishTime: dingtalk.ParseDingTalkTime(detail.ProcessInstance.FinishTime),
			LastSyncedAt:       time.Now(),
		}

		if existing != nil {
			approval.ID = existing.ID
			approval.CreatedAt = existing.CreatedAt
		}

		// Map form component values to specific fields
		s.mapFormValues(approval, detail.ProcessInstance.FormComponentValues)

		// Map operation records to analysis/action fields and build comments
		s.mapOperationRecords(approval, detail.ProcessInstance.OperationRecords, userNameCache)

		if err := s.repo.UpsertApproval(ctx, approval); err != nil {
			s.logger.Error("Failed to upsert approval", zap.Error(err))
			continue
		}

		// Get approval ID (might be new)
		if isNew {
			existing, _ = s.repo.GetByProcessInstanceID(ctx, instanceID)
			if existing != nil {
				approval.ID = existing.ID
			}
			created++
		} else {
			updated++
		}

		// Handle attachments
		s.repo.DeleteAttachments(ctx, approval.ID)
		s.processAttachments(ctx, approval.ID, detail.ProcessInstance.FormComponentValues)

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// Update sync log
	now := time.Now()
	syncLog.Status = "completed"
	syncLog.RecordsProcessed = len(allInstanceIDs)
	syncLog.RecordsCreated = created
	syncLog.RecordsUpdated = updated
	syncLog.CompletedAt = &now
	s.repo.UpdateSyncLog(ctx, syncLog)

	s.logger.Info("Sync completed",
		zap.Int("processed", len(allInstanceIDs)),
		zap.Int("created", created),
		zap.Int("updated", updated))

	return syncLog, nil
}

// mapFormValues maps DingTalk form component values to NCRApproval fields
func (s *Service) mapFormValues(approval *NCRApproval, formValues []dingtalk.FormComponentValue) {
	for _, fv := range formValues {
		fieldName := strings.TrimSpace(fv.Name)
		value := fv.Value

		// Parse multi-select values (JSON arrays) to comma-separated string
		if fv.ComponentType == "DDMultiSelectField" {
			var values []string
			if err := json.Unmarshal([]byte(value), &values); err == nil {
				value = strings.Join(values, ", ")
			}
		}

		// Map by field name
		switch fieldName {
		case "TANGGAL :":
			if t, err := time.Parse("2006-01-02", value); err == nil {
				approval.Tanggal = &t
			}
		case "DITUJUKAN KEPADA :":
			approval.DitujukanKepada = value
		case "DILAPORKAN OLEH :":
			approval.DilaporkanOleh = value
		case "KATEGORI :":
			approval.Kategori = value
		case "NAMA PROJECT :":
			approval.NamaProject = value
		case "NOMOR FPPP : ", "NOMOR FPPP :":
			approval.NomorFPPP = value
		case "NOMOR PRODUCTION ORDER :":
			approval.NomorProductionOrder = value
		case "NAMA  ITEM / PRODUCT :", "NAMA ITEM / PRODUCT :":
			approval.NamaItemProduct = value
		case "DESKRIPSI MASALAH :":
			approval.DeskripsiMasalah = value
		case "TO/TIDAK TO :":
			approval.ToTidakTo = value
		case "URGENT , BUTUH KAPAN : ", "URGENT , BUTUH KAPAN :":
			approval.UrgentButuhKapan = value
		case "CATATAN TAMBAHAN : ", "CATATAN TAMBAHAN :":
			approval.CatatanTambahan = value
		case "DETAIL MATERIAL YANG DIBUTUHKAN :":
			approval.DetailMaterialYangDibutuhkan = value
		}
	}
}

// mapOperationRecords maps operation records to analysis/action fields and builds formatted comments
// Note: DingTalk API does not provide showName in operation_records, so we map EXECUTE_TASK_NORMAL
// operations by order: 1st=analisis, 2nd=nama, 3rd=perbaikan, 4th=pencegahan
func (s *Service) mapOperationRecords(approval *NCRApproval, records []dingtalk.OperationRecord, userNameCache map[string]string) {
	var comments []string
	executeTaskIndex := 0

	for _, op := range records {
		if op.Remark == "" || op.Remark == "-" || op.Remark == "null" {
			continue
		}

		userName := s.client.GetUserName(op.UserID, userNameCache)

		// Format timestamp
		var timeStr string
		if opTime := dingtalk.ParseDingTalkTime(op.Date); opTime != nil {
			timeStr = opTime.Format("2006-01-02 15:04")
		}

		// Debug log for troubleshooting
		s.logger.Debug("Processing operation record",
			zap.String("operation_type", op.OperationType),
			zap.String("remark_preview", op.Remark[:min(50, len(op.Remark))]),
			zap.Int("execute_task_index", executeTaskIndex))

		// Map by operation type
		switch op.OperationType {
		case "EXECUTE_TASK_NORMAL":
			// These are the workflow stage responses
			// Map by order: 1=analisis, 2=nama, 3=perbaikan, 4=pencegahan
			switch executeTaskIndex {
			case 0:
				approval.AnalisisPenyebabMasalah = op.Remark
			case 1:
				approval.NamaYangMelakukanMasalah = op.Remark
			case 2:
				approval.TindakanPerbaikan = op.Remark
			case 3:
				approval.TindakanPencegahan = op.Remark
			default:
				// Additional workflow steps go to comments
				if timeStr != "" {
					comments = append(comments, fmt.Sprintf("(User) %s - %s :\n%s", userName, timeStr, op.Remark))
				} else {
					comments = append(comments, fmt.Sprintf("(User) %s :\n%s", userName, op.Remark))
				}
			}
			executeTaskIndex++
		case "ADD_REMARK":
			// Regular comments go to remark_comment
			if timeStr != "" {
				comments = append(comments, fmt.Sprintf("(User) %s - %s :\n%s", userName, timeStr, op.Remark))
			} else {
				comments = append(comments, fmt.Sprintf("(User) %s :\n%s", userName, op.Remark))
			}
		default:
			// Other operation types with remarks also go to comments
			if op.Remark != "" {
				if timeStr != "" {
					comments = append(comments, fmt.Sprintf("(User) %s - %s :\n%s", userName, timeStr, op.Remark))
				} else {
					comments = append(comments, fmt.Sprintf("(User) %s :\n%s", userName, op.Remark))
				}
			}
		}
	}

	if len(comments) > 0 {
		approval.RemarkComment = strings.Join(comments, "\n\n")
	}
}

// processAttachments extracts and saves attachments from form values
func (s *Service) processAttachments(ctx context.Context, approvalID uuid.UUID, formValues []dingtalk.FormComponentValue) {
	for _, fv := range formValues {
		if fv.ComponentType == "DDPhotoField" {
			var urls []string
			if err := json.Unmarshal([]byte(fv.Value), &urls); err == nil {
				for _, url := range urls {
					s.repo.CreateAttachments(ctx, []NCRAttachment{{
						NCRApprovalID:  approvalID,
						AttachmentType: "photo",
						FieldName:      fv.Name,
						FileURL:        url,
					}})
				}
			}
		} else if fv.ComponentType == "DDAttachment" {
			var attachments []struct {
				SpaceID  string `json:"spaceId"`
				FileName string `json:"fileName"`
				FileSize int64  `json:"fileSize"`
				FileType string `json:"fileType"`
				FileID   string `json:"fileId"`
			}
			if err := json.Unmarshal([]byte(fv.Value), &attachments); err == nil {
				for _, att := range attachments {
					s.repo.CreateAttachments(ctx, []NCRAttachment{{
						NCRApprovalID:  approvalID,
						AttachmentType: "file",
						FieldName:      fv.Name,
						FileName:       att.FileName,
						FileSize:       att.FileSize,
						FileType:       att.FileType,
						SpaceID:        att.SpaceID,
						FileID:         att.FileID,
					}})
				}
			}
		}
	}
}

// ListApprovals lists approvals with filters
func (s *Service) ListApprovals(ctx context.Context, params ListParams) ([]NCRApproval, int64, error) {
	return s.repo.ListApprovals(ctx, params)
}

// GetApproval gets a single approval with details
func (s *Service) GetApproval(ctx context.Context, id uuid.UUID) (*NCRApproval, error) {
	return s.repo.GetApprovalWithDetails(ctx, id)
}

// GetStats gets dashboard statistics (backwards compatible, no filters)
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStatsWithFilters(ctx, StatsParams{})
}

// GetStatsWithFilters gets dashboard statistics with filters
func (s *Service) GetStatsWithFilters(ctx context.Context, params StatsParams) (map[string]interface{}, error) {
	return s.repo.GetStatsWithFilters(ctx, params)
}

// GetFilterOptions gets distinct values for filter dropdowns
func (s *Service) GetFilterOptions(ctx context.Context) (*FilterOptions, error) {
	return s.repo.GetFilterOptions(ctx)
}

// ListSyncLogs lists sync logs
func (s *Service) ListSyncLogs(ctx context.Context, page, pageSize int) ([]SyncLog, int64, error) {
	return s.repo.ListSyncLogs(ctx, page, pageSize)
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
