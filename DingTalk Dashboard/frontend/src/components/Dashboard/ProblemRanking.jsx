import { useState, useEffect } from 'react';
import { dashboardService } from '../../services/dashboardApi';
import './ProblemRanking.css';

const ProblemRanking = ({ filters }) => {
    const [problems, setProblems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [dateRange, setDateRange] = useState({ start: '', end: '' });

    // Get current month date range for default
    const getCurrentMonthRange = () => {
        const now = new Date();
        const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
        const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);

        const formatDate = (date) => date.toISOString().split('T')[0];

        return { start: formatDate(firstDay), end: formatDate(lastDay) };
    };

    useEffect(() => {
        loadProblemRanking();
    }, [
        filters.department, filters.kategori, filters.start_date, filters.end_date,
        filters.ditujukan_kepada, filters.dilaporkan_oleh, filters.status, filters.search
    ]);

    const loadProblemRanking = async () => {
        setLoading(true);
        try {
            const params = {};

            // Date filter - default to current month if not specified
            if (filters.start_date || filters.end_date) {
                if (filters.start_date) params.start_date = filters.start_date;
                if (filters.end_date) params.end_date = filters.end_date;
                setDateRange({ start: filters.start_date, end: filters.end_date });
            } else {
                const currentMonth = getCurrentMonthRange();
                params.start_date = currentMonth.start;
                params.end_date = currentMonth.end;
                setDateRange(currentMonth);
            }

            // All other filters
            if (filters.department) params.department = filters.department;
            if (filters.kategori) params.kategori = filters.kategori;
            if (filters.ditujukan_kepada) params.ditujukan_kepada = filters.ditujukan_kepada;
            if (filters.dilaporkan_oleh) params.dilaporkan_oleh = filters.dilaporkan_oleh;
            if (filters.status) params.status = filters.status;
            if (filters.search) params.search = filters.search;

            const response = await dashboardService.getProblemRanking(params);
            if (response.success) {
                setProblems(response.data || []);
            }
        } catch (error) {
            console.error('Failed to load problem ranking:', error);
        }
        setLoading(false);
    };

    const getRankLabel = (rank) => `#${rank}`;

    const getScoreColor = (score) => {
        if (score >= 70) return 'score-high';
        if (score >= 40) return 'score-medium';
        return 'score-low';
    };

    const formatDateLabel = () => {
        if (!dateRange.start || !dateRange.end) return '';

        const start = new Date(dateRange.start);
        const end = new Date(dateRange.end);

        if (start.getMonth() === end.getMonth() && start.getFullYear() === end.getFullYear()) {
            return start.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
        }

        return `${start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} - ${end.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}`;
    };

    if (loading) {
        return (
            <div className="problem-ranking-container">
                <div className="problem-ranking-header">
                    <h3>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M12 20V10" /><path d="M18 20V4" /><path d="M6 20v-6" />
                        </svg>
                        Top Problem Analysis
                    </h3>
                </div>
                <div className="problem-ranking-loading">
                    <div className="spinner"></div>
                    <span>Analyzing problems...</span>
                </div>
            </div>
        );
    }

    return (
        <div className="problem-ranking-container">
            <div className="problem-ranking-header">
                <h3>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M12 20V10" /><path d="M18 20V4" /><path d="M6 20v-6" />
                    </svg>
                    Top Problem Analysis (RPN Ranking)
                </h3>
                <div className="header-badges">
                    <span className="date-badge">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                            <line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" />
                            <line x1="3" y1="10" x2="21" y2="10" />
                        </svg>
                        {formatDateLabel()}
                    </span>
                    <span className="problem-count">{problems.length} clusters</span>
                </div>
            </div>

            {problems.length === 0 ? (
                <div className="problem-ranking-empty">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                        <circle cx="12" cy="12" r="10" /><path d="M8 12h8M12 8v8" />
                    </svg>
                    <p>No problem data available for this period</p>
                </div>
            ) : (
                <div className="problem-ranking-grid">
                    {problems.map((problem) => (
                        <div key={problem.rank} className="problem-card">
                            <div className="problem-card-header">
                                <span className="problem-rank">{getRankLabel(problem.rank)}</span>
                                <div className="problem-meta">
                                    <span className="frequency-badge">
                                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                                            <circle cx="9" cy="7" r="4" />
                                        </svg>
                                        {problem.frequency}x
                                    </span>
                                    <span className={`rpn-score ${getScoreColor(problem.rpn_score)}`}>
                                        RPN: {problem.rpn_score.toFixed(1)}
                                    </span>
                                </div>
                            </div>
                            <p className="problem-description">
                                {problem.description || 'No description available'}
                            </p>
                            {problem.kategori && (
                                <span className="kategori-tag">{problem.kategori}</span>
                            )}
                        </div>
                    ))}
                </div>
            )}

            <div className="algorithm-info">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" />
                </svg>
                <span>Trigram + LCS Similarity | Weighted RPN | Centroid Selection</span>
            </div>
        </div>
    );
};

export default ProblemRanking;
