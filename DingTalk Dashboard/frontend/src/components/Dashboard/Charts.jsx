import {
    Chart as ChartJS,
    ArcElement,
    CategoryScale,
    LinearScale,
    BarElement,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler,
} from 'chart.js';
import { Doughnut, Bar, Line } from 'react-chartjs-2';
import './Charts.css';

ChartJS.register(
    ArcElement,
    CategoryScale,
    LinearScale,
    BarElement,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
);

const chartColors = {
    primary: '#6366f1',
    secondary: '#818cf8',
    success: '#10b981',
    warning: '#f59e0b',
    error: '#ef4444',
    info: '#3b82f6',
};

// Light theme chart colors
const lightTheme = {
    gridColor: '#e2e8f0',
    textColor: '#64748b',
    tooltipBg: '#ffffff',
    tooltipBorder: '#e2e8f0',
    tooltipText: '#334155',
};

// Format month label from YYYY-MM to readable format
const formatMonthLabel = (monthStr) => {
    if (!monthStr) return '';
    const [year, month] = monthStr.split('-');
    const date = new Date(parseInt(year), parseInt(month) - 1);
    return date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
};

const Charts = ({ stats, dateFilter }) => {
    if (!stats) return null;

    // Check if date filter is active
    const hasDateFilter = dateFilter?.start_date || dateFilter?.end_date;

    // Status distribution data
    const statusData = {
        labels: ['Running', 'Approved', 'Rejected'],
        datasets: [
            {
                data: [stats.running || 0, stats.approved || 0, stats.rejected || 0],
                backgroundColor: [chartColors.warning, chartColors.success, chartColors.error],
                borderWidth: 0,
                hoverOffset: 10,
            },
        ],
    };

    // Department distribution data
    const deptCounts = stats.department_counts || [];
    const deptData = {
        labels: deptCounts.map(d => d.department || 'Unknown').slice(0, 8),
        datasets: [
            {
                label: 'NCR Count',
                data: deptCounts.map(d => d.count).slice(0, 8),
                backgroundColor: chartColors.primary,
                borderRadius: 6,
                maxBarThickness: 40,
            },
        ],
    };

    // Use real trend data from backend
    const rawTrendData = stats.trend_data || [];
    const trendLabels = rawTrendData.map(d => formatMonthLabel(d.month));
    const trendCounts = rawTrendData.map(d => d.count);

    const trendData = {
        labels: trendLabels.length > 0 ? trendLabels : ['No Data'],
        datasets: [
            {
                label: 'NCR Reports',
                data: trendCounts.length > 0 ? trendCounts : [0],
                fill: true,
                backgroundColor: 'rgba(99, 102, 241, 0.1)',
                borderColor: chartColors.primary,
                borderWidth: 2,
                tension: 0.4,
                pointRadius: 4,
                pointHoverRadius: 6,
                pointBackgroundColor: chartColors.primary,
                pointBorderColor: '#ffffff',
                pointBorderWidth: 2,
            },
        ],
    };

    const doughnutOptions = {
        responsive: true,
        maintainAspectRatio: false,
        cutout: '65%',
        plugins: {
            legend: {
                position: 'bottom',
                labels: {
                    color: lightTheme.textColor,
                    padding: 16,
                    usePointStyle: true,
                    pointStyle: 'circle',
                    font: {
                        size: 13,
                        weight: 500,
                    },
                },
            },
            tooltip: {
                backgroundColor: lightTheme.tooltipBg,
                titleColor: lightTheme.tooltipText,
                bodyColor: lightTheme.textColor,
                borderColor: lightTheme.tooltipBorder,
                borderWidth: 1,
                padding: 12,
                cornerRadius: 8,
            },
        },
    };

    const barOptions = {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: 'y',
        plugins: {
            legend: {
                display: false,
            },
            tooltip: {
                backgroundColor: lightTheme.tooltipBg,
                titleColor: lightTheme.tooltipText,
                bodyColor: lightTheme.textColor,
                borderColor: lightTheme.tooltipBorder,
                borderWidth: 1,
                padding: 12,
                cornerRadius: 8,
            },
        },
        scales: {
            x: {
                grid: {
                    color: lightTheme.gridColor,
                },
                ticks: {
                    color: lightTheme.textColor,
                    font: { size: 12 },
                },
            },
            y: {
                grid: {
                    display: false,
                },
                ticks: {
                    color: lightTheme.textColor,
                    font: { size: 12 },
                },
            },
        },
    };

    const lineOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: false,
            },
            tooltip: {
                backgroundColor: lightTheme.tooltipBg,
                titleColor: lightTheme.tooltipText,
                bodyColor: lightTheme.textColor,
                borderColor: lightTheme.tooltipBorder,
                borderWidth: 1,
                padding: 12,
                cornerRadius: 8,
            },
        },
        scales: {
            x: {
                grid: {
                    color: lightTheme.gridColor,
                    drawBorder: false,
                },
                ticks: {
                    color: lightTheme.textColor,
                    font: { size: 11 },
                    maxRotation: 45,
                },
            },
            y: {
                grid: {
                    color: lightTheme.gridColor,
                    drawBorder: false,
                },
                ticks: {
                    color: lightTheme.textColor,
                    font: { size: 12 },
                },
                beginAtZero: true,
            },
        },
    };

    // Calculate date range display text
    const getDateRangeText = () => {
        if (hasDateFilter && dateFilter.start_date && dateFilter.end_date) {
            return `${dateFilter.start_date} - ${dateFilter.end_date}`;
        }
        if (rawTrendData.length > 0) {
            const first = rawTrendData[0]?.month;
            const last = rawTrendData[rawTrendData.length - 1]?.month;
            return `${formatMonthLabel(first)} - ${formatMonthLabel(last)}`;
        }
        return 'All Time';
    };

    return (
        <>
            {/* NCR Trend Chart - Full Width */}
            <div className="chart-card light chart-full-width animate-fade-in">
                <div className="chart-card-header">
                    <h3>NCR Trend Over Time</h3>
                    <span className={`date-filter-badge ${hasDateFilter ? '' : 'default'}`}>
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                            <line x1="16" y1="2" x2="16" y2="6" />
                            <line x1="8" y1="2" x2="8" y2="6" />
                            <line x1="3" y1="10" x2="21" y2="10" />
                        </svg>
                        {getDateRangeText()}
                    </span>
                </div>
                <div className="chart-container chart-trend">
                    <Line data={trendData} options={lineOptions} />
                </div>
            </div>

            <div className="chart-card light animate-fade-in">
                <div className="chart-card-header">
                    <h3>Status Distribution</h3>
                    {hasDateFilter && (
                        <span className="date-filter-badge">Filtered</span>
                    )}
                </div>
                <div className="chart-container chart-doughnut">
                    <Doughnut data={statusData} options={doughnutOptions} />
                </div>
            </div>

            <div className="chart-card light animate-fade-in" style={{ animationDelay: '0.1s' }}>
                <div className="chart-card-header">
                    <h3>By Department</h3>
                    {hasDateFilter && (
                        <span className="date-filter-badge">Filtered</span>
                    )}
                </div>
                <div className="chart-container chart-bar">
                    {deptCounts.length > 0 ? (
                        <Bar data={deptData} options={barOptions} />
                    ) : (
                        <div className="chart-empty">
                            <p className="text-muted">No department data available</p>
                        </div>
                    )}
                </div>
            </div>
        </>
    );
};

export default Charts;
