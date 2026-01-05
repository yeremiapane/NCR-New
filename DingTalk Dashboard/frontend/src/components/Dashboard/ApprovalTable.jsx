import './ApprovalTable.css';

const ApprovalTable = ({ approvals, loading, pagination, onViewDetail, onPageChange }) => {
    const formatDate = (dateStr) => {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleDateString('id-ID', {
            year: 'numeric',
            month: 'short',
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

    const getStatusBadge = (status) => {
        const statusClass = status?.toLowerCase() === 'running' ? 'badge-running' : 'badge-completed';
        return <span className={`badge ${statusClass}`}>{status || '-'}</span>;
    };

    const getResultBadge = (result) => {
        if (!result) return '-';
        const resultClass = result.toLowerCase() === 'agree' ? 'badge-agree' : 'badge-refuse';
        return <span className={`badge ${resultClass}`}>{result}</span>;
    };

    const truncateText = (text, maxLength = 30) => {
        if (!text) return '-';
        return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
    };

    return (
        <div className="approval-table-wrapper">
            <div className="table-scroll-hint">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M14 5l7 7-7 7M3 12h18" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                Scroll horizontally to see all columns
            </div>
            <div className="table-container">
                <table>
                    <thead>
                        <tr>
                            <th className="sticky-col">Action</th>
                            <th>Business ID</th>
                            <th>Title</th>
                            <th>Tanggal</th>
                            <th>Originator</th>
                            <th>Department</th>
                            <th>Kategori</th>
                            <th>Nama Project</th>
                            <th>Nomor FPPP</th>
                            <th>Ditujukan Kepada</th>
                            <th>Dilaporkan Oleh</th>
                            <th>Status</th>
                            <th>Result</th>
                            <th>TO/Tidak TO</th>
                            <th>Temuan</th>
                            <th>Created</th>
                        </tr>
                    </thead>
                    <tbody>
                        {loading ? (
                            <tr>
                                <td colSpan="16" className="table-loading">
                                    <div className="spinner"></div>
                                    <span>Loading...</span>
                                </td>
                            </tr>
                        ) : approvals.length === 0 ? (
                            <tr>
                                <td colSpan="16" className="table-empty">
                                    No approvals found
                                </td>
                            </tr>
                        ) : (
                            approvals.map((approval) => (
                                <tr key={approval.id}>
                                    <td className="sticky-col">
                                        <button
                                            className="btn btn-primary btn-sm"
                                            onClick={() => onViewDetail(approval.id)}
                                        >
                                            View
                                        </button>
                                    </td>
                                    <td className="text-mono">{approval.business_id || '-'}</td>
                                    <td className="table-title" title={approval.title}>
                                        {truncateText(approval.title, 40)}
                                    </td>
                                    <td className="text-sm">{formatDate(approval.tanggal)}</td>
                                    <td>{approval.originator_name || '-'}</td>
                                    <td>{truncateText(approval.originator_dept_name, 25)}</td>
                                    <td>{approval.kategori || '-'}</td>
                                    <td>{truncateText(approval.nama_project, 30)}</td>
                                    <td>{approval.nomor_fppp || '-'}</td>
                                    <td>{truncateText(approval.ditujukan_kepada, 25)}</td>
                                    <td>{truncateText(approval.dilaporkan_oleh, 25)}</td>
                                    <td>{getStatusBadge(approval.status)}</td>
                                    <td>{getResultBadge(approval.result)}</td>
                                    <td>{approval.to_tidak_to || '-'}</td>
                                    <td className="table-temuan" title={approval.temuan}>
                                        {truncateText(approval.temuan, 40)}
                                    </td>
                                    <td className="text-sm text-secondary">
                                        {formatDateTime(approval.dingtalk_create_time)}
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {pagination.total > 0 && (
                <div className="table-pagination">
                    <span className="text-sm text-secondary">
                        Showing {((pagination.page - 1) * pagination.page_size) + 1} to{' '}
                        {Math.min(pagination.page * pagination.page_size, pagination.total)} of{' '}
                        {pagination.total} results
                    </span>
                    <div className="pagination-buttons">
                        <button
                            className="btn btn-secondary btn-sm"
                            disabled={pagination.page <= 1}
                            onClick={() => onPageChange(pagination.page - 1)}
                        >
                            Previous
                        </button>
                        <span className="pagination-current">
                            {pagination.page} / {pagination.total_pages}
                        </span>
                        <button
                            className="btn btn-secondary btn-sm"
                            disabled={pagination.page >= pagination.total_pages}
                            onClick={() => onPageChange(pagination.page + 1)}
                        >
                            Next
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ApprovalTable;
