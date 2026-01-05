import './StatsCards.css';

const StatsCards = ({ stats }) => {
    if (!stats) return null;

    const cards = [
        {
            label: 'Total Approvals',
            value: stats.total || 0,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2" />
                    <rect x="9" y="3" width="6" height="4" rx="1" />
                </svg>
            ),
            color: 'blue',
        },
        {
            label: 'Running',
            value: stats.running || 0,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <polyline points="12 6 12 12 16 14" />
                </svg>
            ),
            color: 'yellow',
        },
        {
            label: 'Approved',
            value: stats.approved || 0,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                    <polyline points="22 4 12 14.01 9 11.01" />
                </svg>
            ),
            color: 'green',
        },
        {
            label: 'Rejected',
            value: stats.rejected || 0,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="15" y1="9" x2="9" y2="15" />
                    <line x1="9" y1="9" x2="15" y2="15" />
                </svg>
            ),
            color: 'red',
        },
    ];

    return (
        <div className="stats-cards">
            {cards.map((card, index) => (
                <div
                    key={card.label}
                    className={`stats-card stats-card-${card.color}`}
                    style={{ animationDelay: `${index * 0.1}s` }}
                >
                    <div className="stats-card-icon">{card.icon}</div>
                    <div className="stats-card-content">
                        <span className="stats-card-value">{card.value.toLocaleString()}</span>
                        <span className="stats-card-label">{card.label}</span>
                    </div>
                </div>
            ))}
        </div>
    );
};

export default StatsCards;
