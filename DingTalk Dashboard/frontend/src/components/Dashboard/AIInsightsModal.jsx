import React, { useState } from 'react';
import './AIInsightsModal.css';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8087';

// SVG Icons
const SparklesIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 3l1.5 4.5L18 9l-4.5 1.5L12 15l-1.5-4.5L6 9l4.5-1.5L12 3z" />
        <path d="M5 19l1 3 1-3 3-1-3-1-1-3-1 3-3 1 3 1z" />
        <path d="M19 13l1 2 1-2 2-1-2-1-1-2-1 2-2 1 2 1z" />
    </svg>
);

const BrainIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 5a3 3 0 1 0-5.997.125 4 4 0 0 0-2.526 5.77 4 4 0 0 0 .556 6.588A4 4 0 1 0 12 18Z" />
        <path d="M12 5a3 3 0 1 1 5.997.125 4 4 0 0 1 2.526 5.77 4 4 0 0 1-.556 6.588A4 4 0 1 1 12 18Z" />
        <path d="M15 13a4.5 4.5 0 0 1-3-4 4.5 4.5 0 0 1-3 4" />
        <path d="M17.599 6.5a3 3 0 0 0 .399-1.375" />
        <path d="M6.003 5.125A3 3 0 0 0 6.401 6.5" />
        <path d="M3.477 10.896a4 4 0 0 1 .585-.396" />
        <path d="M19.938 10.5a4 4 0 0 1 .585.396" />
        <path d="M6 18a4 4 0 0 1-1.967-.516" />
        <path d="M19.967 17.484A4 4 0 0 1 18 18" />
    </svg>
);

const RefreshIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 12a9 9 0 1 1-3-6.71" />
        <path d="M21 3v6h-6" />
    </svg>
);

const CloseIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18" />
        <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
);

const AlertIcon = () => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
    </svg>
);

const AIInsightsModal = ({ isOpen, onClose, filters = {} }) => {
    const [insights, setInsights] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [metadata, setMetadata] = useState(null);

    const generateInsights = async () => {
        setLoading(true);
        setError(null);

        try {
            // Build query params from filters
            const params = new URLSearchParams();
            if (filters.startDate) params.append('start_date', filters.startDate);
            if (filters.endDate) params.append('end_date', filters.endDate);
            if (filters.department) params.append('department', filters.department);
            if (filters.ditujukanKepada) params.append('ditujukan_kepada', filters.ditujukanKepada);
            if (filters.dilaporkanOleh) params.append('dilaporkan_oleh', filters.dilaporkanOleh);
            if (filters.kategori) params.append('kategori', filters.kategori);
            if (filters.status) params.append('status', filters.status);

            // Get auth token (key matches AuthContext)
            const token = localStorage.getItem('access_token');

            // Build headers - only include Authorization if token exists
            const headers = {
                'Content-Type': 'application/json',
            };
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }

            const response = await fetch(`${API_BASE_URL}/api/v1/ai/insights?${params}`, {
                method: 'GET',
                headers: headers,
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.error || data.message || 'Failed to generate insights');
            }

            setInsights(data.data.insights);
            setMetadata({
                model: data.data.model,
                processTime: data.data.process_time_seconds,
                generatedAt: new Date(data.data.generated_at),
            });
        } catch (err) {
            console.error('AI Insights error:', err);
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleOverlayClick = (e) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    if (!isOpen) return null;

    const renderSkeletons = () => (
        <div className="insights-skeleton">
            {[1, 2, 3, 4].map((i) => (
                <div key={i} className="skeleton-card">
                    <div className="skeleton-line short"></div>
                    <div className="skeleton-line medium"></div>
                    <div className="skeleton-line long"></div>
                </div>
            ))}
        </div>
    );

    const renderEmptyState = () => (
        <div className="ai-empty-state">
            <div className="ai-empty-icon">
                <BrainIcon />
            </div>
            <h4>AI Insights Ready</h4>
            <p>Click "Generate Insights" to analyze your NCR data with AI</p>
        </div>
    );

    const renderError = () => (
        <div className="ai-error-state">
            <h4><AlertIcon /> Unable to Generate Insights</h4>
            <p>{error}</p>
            <p style={{ marginTop: '12px', fontSize: '0.8rem', color: '#64748b' }}>
                Make sure Ollama is running: <code>ollama serve</code>
            </p>
        </div>
    );

    const renderInsights = () => (
        <div className="insights-grid">
            {insights.map((insight, index) => (
                <div
                    key={index}
                    className={`insight-card severity-${insight.severity}`}
                >
                    <div className="insight-header">
                        <span className={`insight-type-badge ${insight.type}`}>
                            {insight.type}
                        </span>
                        <div className={`severity-indicator ${insight.severity}`}></div>
                    </div>
                    <h4 className="insight-title">{insight.title}</h4>
                    <p className="insight-description">{insight.description}</p>
                </div>
            ))}
        </div>
    );

    return (
        <div className="ai-modal-overlay" onClick={handleOverlayClick}>
            <div className="ai-modal">
                <div className="ai-modal-header">
                    <div className="ai-modal-title">
                        <div className="ai-icon-wrapper">
                            <BrainIcon />
                        </div>
                        <h2>AI Insights</h2>
                    </div>

                    <div className="ai-modal-actions">
                        {metadata && (
                            <div className="ai-meta-info">
                                <span className="ai-model-badge">{metadata.model}</span>
                                <span className="ai-process-time">
                                    {metadata.processTime.toFixed(1)}s
                                </span>
                            </div>
                        )}

                        <button
                            className="generate-btn"
                            onClick={generateInsights}
                            disabled={loading}
                        >
                            {loading ? (
                                <>
                                    <span className="spinner"></span>
                                    Analyzing...
                                </>
                            ) : (
                                <>
                                    {insights ? <RefreshIcon /> : <SparklesIcon />}
                                    {insights ? 'Regenerate' : 'Generate Insights'}
                                </>
                            )}
                        </button>

                        <button className="close-btn" onClick={onClose}>
                            <CloseIcon />
                        </button>
                    </div>
                </div>

                <div className="ai-modal-body">
                    {loading && renderSkeletons()}
                    {!loading && error && renderError()}
                    {!loading && !error && !insights && renderEmptyState()}
                    {!loading && !error && insights && insights.length > 0 && renderInsights()}
                </div>
            </div>
        </div>
    );
};

// Export the header button component as well
export const AIHeaderButton = ({ onClick }) => (
    <button className="ai-header-btn" onClick={onClick}>
        <SparklesIcon />
        AI Insights
    </button>
);

export default AIInsightsModal;
