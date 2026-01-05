package dingtalk

import "time"

// ApprovalListResponse represents the response from list instance IDs API
type ApprovalListResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Result  struct {
		List       []string `json:"list"`
		NextCursor int64    `json:"next_cursor"`
	} `json:"result"`
}

// ApprovalDetailResponse represents the response from get instance detail API
// Note: The response uses "process_instance" not "result" and snake_case field names
type ApprovalDetailResponse struct {
	ErrCode         int              `json:"errcode"`
	ErrMsg          string           `json:"errmsg"`
	ProcessInstance *ProcessInstance `json:"process_instance"`
}

// ProcessInstance represents the actual approval instance data
type ProcessInstance struct {
	Title               string               `json:"title"`
	Status              string               `json:"status"`
	Result              string               `json:"result"`
	BusinessID          string               `json:"business_id"`
	OriginatorUserID    string               `json:"originator_userid"`
	OriginatorDeptID    string               `json:"originator_dept_id"`
	OriginatorDeptName  string               `json:"originator_dept_name"`
	CreateTime          string               `json:"create_time"`
	FinishTime          string               `json:"finish_time"`
	FormComponentValues []FormComponentValue `json:"form_component_values"`
	OperationRecords    []OperationRecord    `json:"operation_records"`
	Tasks               []Task               `json:"tasks"`
	CCUserIDs           []string             `json:"cc_userids"`
}

// FormComponentValue represents a form field value
type FormComponentValue struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Value         string `json:"value"`
	ExtValue      string `json:"ext_value"`
	ComponentType string `json:"component_type"`
	BizAlias      string `json:"biz_alias"`
}

// OperationRecordAttachment represents an attachment in operation record
type OperationRecordAttachment struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
	FileType string `json:"file_type"`
}

// OperationRecord represents an operation/comment record
// Note: The API returns snake_case fields: operation_type, operation_result, etc.
type OperationRecord struct {
	UserID          string                      `json:"userid"`
	Date            string                      `json:"date"`
	OperationType   string                      `json:"operation_type"`
	OperationResult string                      `json:"operation_result"`
	Remark          string                      `json:"remark"`
	Attachments     []OperationRecordAttachment `json:"attachments"`
	Images          []string                    `json:"images"`
}

// Task represents a task in the approval flow
type Task struct {
	TaskID     interface{} `json:"taskid"` // Can be string or int64
	UserID     string      `json:"userid"`
	Status     string      `json:"task_status"`
	Result     string      `json:"task_result"`
	CreateTime string      `json:"create_time"`
	FinishTime string      `json:"finish_time"`
	ActivityID string      `json:"activity_id"`
}

// UserInfoResponse represents the response from user info API
type UserInfoResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Result  struct {
		UserID string `json:"userid"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
	} `json:"result"`
}

// ParseDingTalkTime parses DingTalk time format
func ParseDingTalkTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}

	// DingTalk time format: "2026-01-05 11:40:18"
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04Z",
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, timeStr)
		if err == nil {
			return &t
		}
	}

	return nil
}
