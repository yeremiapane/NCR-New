import './ApprovalDetail.css';

const ApprovalDetail = ({ approval, onClose }) => {
    if (!approval) return null;

    const formatDate = (dateStr) => {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleDateString('id-ID', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });
    };

    const formatDateTime = (dateStr) => {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleDateString('id-ID', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const renderField = (label, value) => {
        if (!value) return null;
        return (
            <div className="form-value-item">
                <span className="form-value-label">{label}</span>
                <span className="form-value-content">{value}</span>
            </div>
        );
    };

    const renderMultilineField = (label, value) => {
        if (!value) return null;
        return (
            <div className="form-value-item">
                <span className="form-value-label">{label}</span>
                <pre className="form-value-content form-value-pre">{value}</pre>
            </div>
        );
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>{approval.title || 'NCR Detail'}</h2>
                    <button className="modal-close" onClick={onClose}>
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                        </svg>
                    </button>
                </div>

                <div className="modal-body">
                    {/* Status Info */}
                    <section className="detail-section">
                        <h3>Status</h3>
                        <div className="detail-grid">
                            <div className="detail-item">
                                <span className="detail-label">Status</span>
                                <span className={`badge badge-${approval.status?.toLowerCase()}`}>
                                    {approval.status}
                                </span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">Result</span>
                                <span className={`badge badge-${approval.result?.toLowerCase()}`}>
                                    {approval.result || '-'}
                                </span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">TO/TIDAK TO</span>
                                <span className={`badge ${approval.to_tidak_to?.includes('TIDAK') ? 'badge-refuse' : 'badge-agree'}`}>
                                    {approval.to_tidak_to || '-'}
                                </span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">Created</span>
                                <span>{formatDateTime(approval.dingtalk_create_time)}</span>
                            </div>
                        </div>
                    </section>

                    {/* Originator Info */}
                    <section className="detail-section">
                        <h3>Reported By</h3>
                        <div className="detail-grid">
                            <div className="detail-item">
                                <span className="detail-label">Name</span>
                                <span>{approval.originator_name || '-'}</span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">Department</span>
                                <span>{approval.originator_dept_name || '-'}</span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">Dilaporkan Oleh</span>
                                <span>{approval.dilaporkan_oleh || '-'}</span>
                            </div>
                            <div className="detail-item">
                                <span className="detail-label">Ditujukan Kepada</span>
                                <span>{approval.ditujukan_kepada || '-'}</span>
                            </div>
                        </div>
                    </section>

                    {/* NCR Details */}
                    <section className="detail-section">
                        <h3>NCR Information</h3>
                        <div className="form-values">
                            {renderField('Tanggal', formatDate(approval.tanggal))}
                            {renderField('Kategori', approval.kategori)}
                            {renderField('Nama Project', approval.nama_project)}
                            {renderField('Nomor FPPP', approval.nomor_fppp)}
                            {renderField('Nomor Production Order', approval.nomor_production_order)}
                            {renderField('Nama Item/Product', approval.nama_item_product)}
                            {renderMultilineField('Deskripsi Masalah', approval.deskripsi_masalah)}
                            {renderField('Urgent/Butuh Kapan', approval.urgent_butuh_kapan)}
                            {renderMultilineField('Catatan Tambahan', approval.catatan_tambahan)}
                            {renderMultilineField('Detail Material yang Dibutuhkan', approval.detail_material_yang_dibutuhkan)}
                        </div>
                    </section>

                    {/* Analysis & Actions */}
                    {(approval.analisis_penyebab_masalah || approval.nama_yang_melakukan_masalah ||
                        approval.tindakan_perbaikan || approval.tindakan_pencegahan) && (
                            <section className="detail-section">
                                <h3>Analysis & Corrective Actions</h3>
                                <div className="form-values">
                                    {renderMultilineField('Analisis Penyebab Masalah', approval.analisis_penyebab_masalah)}
                                    {renderField('Nama yang Melakukan Kesalahan', approval.nama_yang_melakukan_masalah)}
                                    {renderMultilineField('Tindakan Perbaikan', approval.tindakan_perbaikan)}
                                    {renderMultilineField('Tindakan Pencegahan', approval.tindakan_pencegahan)}
                                </div>
                            </section>
                        )}

                    {/* Comments */}
                    {approval.remark_comment && (
                        <section className="detail-section">
                            <h3>Comments</h3>
                            <div className="comments-box">
                                <pre className="comments-content">{approval.remark_comment}</pre>
                            </div>
                        </section>
                    )}

                    {/* Attachments */}
                    {approval.attachments && approval.attachments.length > 0 && (
                        <section className="detail-section">
                            <h3>Attachments</h3>
                            <div className="attachments-grid">
                                {approval.attachments.map((att, index) => (
                                    <div key={index} className="attachment-item">
                                        {att.attachment_type === 'photo' && att.file_url ? (
                                            <a href={att.file_url} target="_blank" rel="noopener noreferrer">
                                                <img src={att.file_url} alt={att.field_name} loading="lazy" />
                                            </a>
                                        ) : (
                                            <div className="attachment-file">
                                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                                    <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
                                                    <polyline points="13 2 13 9 20 9" />
                                                </svg>
                                                <span>{att.file_name || 'File'}</span>
                                                {att.file_type && <span className="file-type">.{att.file_type}</span>}
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </section>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ApprovalDetail;
