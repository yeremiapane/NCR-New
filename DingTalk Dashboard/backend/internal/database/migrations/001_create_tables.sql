-- DingTalk NCR Dashboard Database Schema
-- Migration 001: Create tables with specific NCR form fields

-- Main NCR approval instance record with specific columns
CREATE TABLE IF NOT EXISTS ncr_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    process_instance_id VARCHAR(100) UNIQUE NOT NULL,
    business_id VARCHAR(100),
    title VARCHAR(500),
    status VARCHAR(50) NOT NULL,  -- RUNNING, COMPLETED, CANCELED
    result VARCHAR(50),           -- agree, refuse
    
    -- Originator info
    originator_user_id VARCHAR(100),
    originator_name VARCHAR(200),
    originator_dept_id VARCHAR(100),
    originator_dept_name VARCHAR(200),
    
    -- NCR Form specific fields
    tanggal DATE,
    ditujukan_kepada TEXT,          -- Could be multiple values comma-separated
    dilaporkan_oleh TEXT,           -- Could be multiple values comma-separated
    kategori TEXT,                  -- Could be multiple values comma-separated
    nama_project VARCHAR(500),
    nomor_fppp VARCHAR(200),
    nomor_production_order VARCHAR(200),
    nama_item_product TEXT,
    deskripsi_masalah TEXT,
    to_tidak_to VARCHAR(50),
    urgent_butuh_kapan TEXT,
    catatan_tambahan TEXT,
    detail_material_yang_dibutuhkan TEXT,
    
    -- Analysis and corrective action fields (from workflow stages)
    analisis_penyebab_masalah TEXT,
    nama_yang_melakukan_masalah TEXT,
    tindakan_perbaikan TEXT,
    tindakan_pencegahan TEXT,
    
    -- Formatted comments (all remarks combined)
    remark_comment TEXT,
    
    -- Timestamps from DingTalk
    dingtalk_create_time TIMESTAMPTZ,
    dingtalk_finish_time TIMESTAMPTZ,
    
    -- Local timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

-- Attachments and photos (kept separate for flexibility)
CREATE TABLE IF NOT EXISTS ncr_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ncr_approval_id UUID REFERENCES ncr_approvals(id) ON DELETE CASCADE,
    attachment_type VARCHAR(50),  -- photo, file
    field_name VARCHAR(300),      -- Which form field this belongs to
    file_url TEXT,
    file_name VARCHAR(500),
    file_size BIGINT,
    file_type VARCHAR(50),
    space_id VARCHAR(100),
    file_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sync logs for monitoring
CREATE TABLE IF NOT EXISTS sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_type VARCHAR(50) NOT NULL,    -- scheduled, manual
    status VARCHAR(50) NOT NULL,       -- started, completed, failed
    records_processed INT DEFAULT 0,
    records_created INT DEFAULT 0,
    records_updated INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_status ON ncr_approvals(status);
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_created ON ncr_approvals(dingtalk_create_time);
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_tanggal ON ncr_approvals(tanggal);
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_originator ON ncr_approvals(originator_dept_name);
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_kategori ON ncr_approvals(kategori);
CREATE INDEX IF NOT EXISTS idx_ncr_approvals_to_tidak_to ON ncr_approvals(to_tidak_to);
CREATE INDEX IF NOT EXISTS idx_ncr_attachments_approval ON ncr_attachments(ncr_approval_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_started ON sync_logs(started_at);
