import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { dashboardService } from '../services/dashboardApi';
import StatsCards from '../components/Dashboard/StatsCards';
import ApprovalTable from '../components/Dashboard/ApprovalTable';
import Charts from '../components/Dashboard/Charts';
import FilterPanel from '../components/Dashboard/FilterPanel';
import ProblemRanking from '../components/Dashboard/ProblemRanking';
import WordCloud from '../components/Dashboard/WordCloud';
import ApprovalDetail from '../components/Dashboard/ApprovalDetail';
import AIInsightsModal, { AIHeaderButton } from '../components/Dashboard/AIInsightsModal';
import './Dashboard.css';

const Dashboard = () => {
    const { user, logout } = useAuth();
    const [stats, setStats] = useState(null);
    const [approvals, setApprovals] = useState([]);
    const [filterOptions, setFilterOptions] = useState(null);
    const [pagination, setPagination] = useState({ page: 1, total: 0, totalPages: 0 });
    const [loading, setLoading] = useState(true);
    const [syncing, setSyncing] = useState(false);
    const [selectedApproval, setSelectedApproval] = useState(null);
    const [aiModalOpen, setAiModalOpen] = useState(false);

    // Get current month date range
    const getCurrentMonthRange = () => {
        const now = new Date();
        const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
        const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        const formatDate = (date) => date.toISOString().split('T')[0];
        return { start: formatDate(firstDay), end: formatDate(lastDay) };
    };

    const currentMonth = getCurrentMonthRange();

    const [filters, setFilters] = useState({
        search: '',
        business_id: '',
        status: '',
        department: '',
        ditujukan_kepada: '',
        dilaporkan_oleh: '',
        kategori: '',
        start_date: currentMonth.start,  // Default to current month
        end_date: currentMonth.end,      // Default to current month
        page: 1,
        page_size: 10,
    });

    // Load filter options on mount
    useEffect(() => {
        loadFilterOptions();
    }, []);

    // Load data when filters change
    useEffect(() => {
        loadData();
    }, [filters]);

    const loadFilterOptions = async () => {
        try {
            const response = await dashboardService.getFilterOptions();
            if (response.success) {
                setFilterOptions(response.data);
            }
        } catch (error) {
            console.error('Failed to load filter options:', error);
        }
    };

    const loadData = async () => {
        setLoading(true);
        try {
            // Use current month as default if no date filter set
            const effectiveStartDate = filters.start_date || currentMonth.start;
            const effectiveEndDate = filters.end_date || currentMonth.end;

            // Build stats filter params (for charts to respond to ALL filters)
            const statsParams = {
                start_date: effectiveStartDate,
                end_date: effectiveEndDate,
            };
            if (filters.search) statsParams.search = filters.search;
            if (filters.status) statsParams.status = filters.status;
            if (filters.department) statsParams.department = filters.department;
            if (filters.ditujukan_kepada) statsParams.ditujukan_kepada = filters.ditujukan_kepada;
            if (filters.dilaporkan_oleh) statsParams.dilaporkan_oleh = filters.dilaporkan_oleh;
            if (filters.kategori) statsParams.kategori = filters.kategori;

            const [statsRes, approvalsRes] = await Promise.all([
                dashboardService.getStats(statsParams),
                dashboardService.getApprovals(filters),
            ]);

            if (statsRes.success) {
                setStats(statsRes.data);
            }

            if (approvalsRes.success) {
                setApprovals(approvalsRes.data.approvals || []);
                setPagination(approvalsRes.data.pagination || {});
            }
        } catch (error) {
            console.error('Failed to load data:', error);
        }
        setLoading(false);
    };

    const handleSync = async () => {
        setSyncing(true);
        try {
            await dashboardService.triggerSync();
            await loadData();
            await loadFilterOptions(); // Refresh filter options after sync
        } catch (error) {
            console.error('Sync failed:', error);
        }
        setSyncing(false);
    };

    const handleViewDetail = async (id) => {
        try {
            const response = await dashboardService.getApproval(id);
            if (response.success) {
                setSelectedApproval(response.data);
            }
        } catch (error) {
            console.error('Failed to load approval:', error);
        }
    };

    const handleFilterChange = (key, value) => {
        setFilters(prev => ({ ...prev, [key]: value, page: 1 }));
    };

    const handlePageChange = (newPage) => {
        setFilters(prev => ({ ...prev, page: newPage }));
    };

    const [exporting, setExporting] = useState(false);
    const handleExport = async () => {
        setExporting(true);
        try {
            // Use effective dates (default to current month)
            const exportParams = {
                start_date: filters.start_date || currentMonth.start,
                end_date: filters.end_date || currentMonth.end,
            };
            if (filters.search) exportParams.search = filters.search;
            if (filters.status) exportParams.status = filters.status;
            if (filters.department) exportParams.department = filters.department;
            if (filters.ditujukan_kepada) exportParams.ditujukan_kepada = filters.ditujukan_kepada;
            if (filters.dilaporkan_oleh) exportParams.dilaporkan_oleh = filters.dilaporkan_oleh;
            if (filters.kategori) exportParams.kategori = filters.kategori;

            const blob = await dashboardService.exportToExcel(exportParams);

            // Create download link
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `NCR_Export_${new Date().toISOString().split('T')[0]}.xlsx`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Export failed:', error);
            alert('Export failed. Please try again.');
        }
        setExporting(false);
    };

    return (
        <div className="dashboard light-theme">
            <header className="dashboard-header">
                <div className="dashboard-header-left">
                    <div className="dashboard-logo">
                        <svg width="32" height="32" viewBox="0 0 40 40" fill="none">
                            <rect width="40" height="40" rx="10" fill="url(#logo-grad)" />
                            <path d="M12 20L18 26L28 14" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
                            <defs>
                                <linearGradient id="logo-grad" x1="0" y1="0" x2="40" y2="40">
                                    <stop stopColor="#6366f1" />
                                    <stop offset="1" stopColor="#818cf8" />
                                </linearGradient>
                            </defs>
                        </svg>
                    </div>
                    <div>
                        <h1>NCR Internal Dashboard</h1>
                        <p className="text-secondary text-sm">Approval workflow monitoring</p>
                    </div>
                </div>
                <div className="dashboard-header-right">
                    <button className="btn btn-primary" onClick={handleExport} disabled={exporting}>
                        {exporting ? (
                            <>
                                <span className="spinner"></span>
                                Exporting...
                            </>
                        ) : (
                            <>
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                    <polyline points="7 10 12 15 17 10" />
                                    <line x1="12" y1="15" x2="12" y2="3" />
                                </svg>
                                Export Excel
                            </>
                        )}
                    </button>
                    <button className="btn btn-secondary" onClick={handleSync} disabled={syncing}>
                        {syncing ? (
                            <>
                                <span className="spinner"></span>
                                Syncing...
                            </>
                        ) : (
                            <>
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path d="M21 12a9 9 0 1 1-3-6.71" strokeLinecap="round" />
                                    <path d="M21 3v6h-6" strokeLinecap="round" strokeLinejoin="round" />
                                </svg>
                                Sync Now
                            </>
                        )}
                    </button>
                    <AIHeaderButton onClick={() => setAiModalOpen(true)} />
                    <div className="user-menu">
                        <span className="text-secondary text-sm">{user?.email}</span>
                        <button className="btn btn-secondary btn-sm" onClick={logout}>
                            Logout
                        </button>
                    </div>
                </div>
            </header>

            <main className="dashboard-content">
                {loading && !stats ? (
                    <div className="dashboard-loading">
                        <div className="spinner"></div>
                        <p>Loading dashboard...</p>
                    </div>
                ) : (
                    <>
                        <StatsCards stats={stats} />

                        {/* Filter Panel */}
                        <FilterPanel
                            filters={filters}
                            onFilterChange={handleFilterChange}
                            filterOptions={filterOptions}
                            loading={loading}
                        />

                        <div className="dashboard-grid">
                            <Charts
                                stats={stats}
                                dateFilter={{
                                    start_date: filters.start_date,
                                    end_date: filters.end_date
                                }}
                            />
                            <WordCloud filters={filters} />
                        </div>

                        {/* Problem Ranking Section */}
                        <ProblemRanking filters={filters} />

                        <section className="dashboard-section">
                            <div className="section-header">
                                <h2>NCR Approvals</h2>
                                <div className="section-info">
                                    <span className="text-secondary">
                                        Showing {approvals.length} of {pagination.total || 0} records
                                    </span>
                                </div>
                            </div>

                            <ApprovalTable
                                approvals={approvals}
                                loading={loading}
                                pagination={pagination}
                                onViewDetail={handleViewDetail}
                                onPageChange={handlePageChange}
                            />
                        </section>
                    </>
                )}
            </main>

            {selectedApproval && (
                <ApprovalDetail
                    approval={selectedApproval}
                    onClose={() => setSelectedApproval(null)}
                />
            )}

            {/* AI Insights Modal */}
            <AIInsightsModal
                isOpen={aiModalOpen}
                onClose={() => setAiModalOpen(false)}
                filters={{
                    startDate: filters.start_date,
                    endDate: filters.end_date,
                    department: filters.department,
                    ditujukanKepada: filters.ditujukan_kepada,
                    dilaporkanOleh: filters.dilaporkan_oleh,
                    kategori: filters.kategori,
                    status: filters.status,
                }}
            />
        </div>
    );
};

export default Dashboard;
