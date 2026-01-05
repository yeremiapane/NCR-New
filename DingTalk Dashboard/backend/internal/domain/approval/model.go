package approval

import (
	"time"

	"github.com/google/uuid"
)

// NCRApproval represents an NCR approval workflow instance with specific fields
type NCRApproval struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProcessInstanceID string    `gorm:"uniqueIndex;size:100;not null" json:"process_instance_id"`
	BusinessID        string    `gorm:"size:100" json:"business_id"`
	Title             string    `gorm:"size:500" json:"title"`
	Status            string    `gorm:"size:50;not null" json:"status"`
	Result            string    `gorm:"size:50" json:"result"`

	// Originator info
	OriginatorUserID   string `gorm:"size:100" json:"originator_user_id"`
	OriginatorName     string `gorm:"size:200" json:"originator_name"`
	OriginatorDeptID   string `gorm:"size:100" json:"originator_dept_id"`
	OriginatorDeptName string `gorm:"size:200" json:"originator_dept_name"`

	// NCR Form specific fields
	Tanggal                      *time.Time `gorm:"type:date" json:"tanggal"`
	DitujukanKepada              string     `gorm:"column:ditujukan_kepada;type:text" json:"ditujukan_kepada"`
	DilaporkanOleh               string     `gorm:"column:dilaporkan_oleh;type:text" json:"dilaporkan_oleh"`
	Kategori                     string     `gorm:"column:kategori;type:text" json:"kategori"`
	NamaProject                  string     `gorm:"column:nama_project;size:500" json:"nama_project"`
	NomorFPPP                    string     `gorm:"column:nomor_fppp;size:200" json:"nomor_fppp"`
	NomorProductionOrder         string     `gorm:"column:nomor_production_order;size:200" json:"nomor_production_order"`
	NamaItemProduct              string     `gorm:"column:nama_item_product;type:text" json:"nama_item_product"`
	DeskripsiMasalah             string     `gorm:"column:deskripsi_masalah;type:text" json:"deskripsi_masalah"`
	ToTidakTo                    string     `gorm:"column:to_tidak_to;size:50" json:"to_tidak_to"`
	UrgentButuhKapan             string     `gorm:"column:urgent_butuh_kapan;type:text" json:"urgent_butuh_kapan"`
	CatatanTambahan              string     `gorm:"column:catatan_tambahan;type:text" json:"catatan_tambahan"`
	DetailMaterialYangDibutuhkan string     `gorm:"column:detail_material_yang_dibutuhkan;type:text" json:"detail_material_yang_dibutuhkan"`

	// Analysis and corrective action fields
	AnalisisPenyebabMasalah  string `gorm:"column:analisis_penyebab_masalah;type:text" json:"analisis_penyebab_masalah"`
	NamaYangMelakukanMasalah string `gorm:"column:nama_yang_melakukan_masalah;type:text" json:"nama_yang_melakukan_masalah"`
	TindakanPerbaikan        string `gorm:"column:tindakan_perbaikan;type:text" json:"tindakan_perbaikan"`
	TindakanPencegahan       string `gorm:"column:tindakan_pencegahan;type:text" json:"tindakan_pencegahan"`

	// Formatted comments (all remarks combined)
	RemarkComment string `gorm:"column:remark_comment;type:text" json:"remark_comment"`

	// Timestamps from DingTalk
	DingTalkCreateTime *time.Time `gorm:"column:dingtalk_create_time" json:"dingtalk_create_time"`
	DingTalkFinishTime *time.Time `gorm:"column:dingtalk_finish_time" json:"dingtalk_finish_time"`

	// Local timestamps
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	LastSyncedAt time.Time `json:"last_synced_at"`

	// Relations
	Attachments []NCRAttachment `gorm:"foreignKey:NCRApprovalID" json:"attachments,omitempty"`
}

func (NCRApproval) TableName() string {
	return "ncr_approvals"
}

// NCRAttachment represents an attachment or photo
type NCRAttachment struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NCRApprovalID  uuid.UUID `gorm:"type:uuid;index" json:"ncr_approval_id"`
	AttachmentType string    `gorm:"size:50" json:"attachment_type"`
	FieldName      string    `gorm:"size:300" json:"field_name"`
	FileURL        string    `gorm:"type:text" json:"file_url"`
	FileName       string    `gorm:"size:500" json:"file_name"`
	FileSize       int64     `json:"file_size"`
	FileType       string    `gorm:"size:50" json:"file_type"`
	SpaceID        string    `gorm:"size:100" json:"space_id"`
	FileID         string    `gorm:"size:100" json:"file_id"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (NCRAttachment) TableName() string {
	return "ncr_attachments"
}

// SyncLog represents a sync operation log entry
type SyncLog struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SyncType         string     `gorm:"size:50;not null" json:"sync_type"`
	Status           string     `gorm:"size:50;not null" json:"status"`
	RecordsProcessed int        `gorm:"default:0" json:"records_processed"`
	RecordsCreated   int        `gorm:"default:0" json:"records_created"`
	RecordsUpdated   int        `gorm:"default:0" json:"records_updated"`
	ErrorMessage     string     `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt        time.Time  `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func (SyncLog) TableName() string {
	return "sync_logs"
}

// Field name mappings from DingTalk form to database columns
var FieldNameMapping = map[string]string{
	"TANGGAL :":                         "tanggal",
	"DITUJUKAN KEPADA :":                "ditujukan_kepada",
	"DILAPORKAN OLEH :":                 "dilaporkan_oleh",
	"KATEGORI :":                        "kategori",
	"NAMA PROJECT :":                    "nama_project",
	"NOMOR FPPP : ":                     "nomor_fppp",
	"NOMOR FPPP :":                      "nomor_fppp",
	"NOMOR PRODUCTION ORDER :":          "nomor_production_order",
	"NAMA  ITEM / PRODUCT :":            "nama_item_product",
	"NAMA ITEM / PRODUCT :":             "nama_item_product",
	"DESKRIPSI MASALAH :":               "deskripsi_masalah",
	"TO/TIDAK TO :":                     "to_tidak_to",
	"URGENT , BUTUH KAPAN : ":           "urgent_butuh_kapan",
	"URGENT , BUTUH KAPAN :":            "urgent_butuh_kapan",
	"CATATAN TAMBAHAN : ":               "catatan_tambahan",
	"CATATAN TAMBAHAN :":                "catatan_tambahan",
	"DETAIL MATERIAL YANG DIBUTUHKAN :": "detail_material_yang_dibutuhkan",
	"ANALISA PENYEBAB MASALAH :":        "analisis_penyebab_masalah",
	"NAMA YANG MELAKUKAN KESALAHAN :":   "nama_yang_melakukan_masalah",
	"TINDAKAN PERBAIKAN :":              "tindakan_perbaikan",
	"TINDAKAN PENCEGAHAN :":             "tindakan_pencegahan",
}

// Workflow stage mappings (showName from operation records)
var WorkflowStageMapping = map[string]string{
	"ANALISA PENYEBAB MASALAH :":      "analisis_penyebab_masalah",
	"NAMA YANG MELAKUKAN KESALAHAN :": "nama_yang_melakukan_masalah",
	"TINDAKAN PERBAIKAN :":            "tindakan_perbaikan",
	"TINDAKAN PENCEGAHAN :":           "tindakan_pencegahan",
}
