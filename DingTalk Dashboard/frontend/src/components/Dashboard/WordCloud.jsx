import { useState, useEffect } from 'react';
import { dashboardService } from '../../services/dashboardApi';
import './WordCloud.css';

const WordCloud = ({ filters }) => {
    const [words, setWords] = useState([]);
    const [loading, setLoading] = useState(true);

    // Get current month date range for default
    const getCurrentMonthRange = () => {
        const now = new Date();
        const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
        const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);

        const formatDate = (date) => date.toISOString().split('T')[0];

        return { start: formatDate(firstDay), end: formatDate(lastDay) };
    };

    useEffect(() => {
        loadWordCloud();
    }, [
        filters.department, filters.kategori, filters.start_date, filters.end_date,
        filters.ditujukan_kepada, filters.dilaporkan_oleh, filters.status, filters.search
    ]);

    const loadWordCloud = async () => {
        setLoading(true);
        try {
            const params = {};

            // Date filter - default to current month
            if (filters.start_date || filters.end_date) {
                if (filters.start_date) params.start_date = filters.start_date;
                if (filters.end_date) params.end_date = filters.end_date;
            } else {
                const currentMonth = getCurrentMonthRange();
                params.start_date = currentMonth.start;
                params.end_date = currentMonth.end;
            }

            // All other filters
            if (filters.department) params.department = filters.department;
            if (filters.kategori) params.kategori = filters.kategori;
            if (filters.ditujukan_kepada) params.ditujukan_kepada = filters.ditujukan_kepada;
            if (filters.dilaporkan_oleh) params.dilaporkan_oleh = filters.dilaporkan_oleh;
            if (filters.status) params.status = filters.status;
            if (filters.search) params.search = filters.search;

            const response = await dashboardService.getWordCloud(params);
            if (response.success) {
                setWords(response.data || []);
            }
        } catch (error) {
            console.error('Failed to load word cloud:', error);
        }
        setLoading(false);
    };

    // Calculate size class based on word frequency
    const getSizeClass = (count, maxCount) => {
        const ratio = count / maxCount;
        if (ratio >= 0.8) return 'word-xl';
        if (ratio >= 0.6) return 'word-lg';
        if (ratio >= 0.4) return 'word-md';
        if (ratio >= 0.2) return 'word-sm';
        return 'word-xs';
    };

    // Get unique gradient for each word
    const getGradientStyle = (index) => {
        const gradients = [
            'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
            'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
            'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
            'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
            'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)',
            'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
            'linear-gradient(135deg, #fbc2eb 0%, #a6c1ee 100%)',
        ];
        return gradients[index % gradients.length];
    };

    const maxCount = words.length > 0 ? words[0].count : 1;

    if (loading) {
        return (
            <div className="chart-card light animate-fade-in word-cloud-card">
                <div className="chart-card-header">
                    <h3>Problem Keywords</h3>
                </div>
                <div className="word-cloud-loading">
                    <div className="spinner"></div>
                    <span>Extracting keywords...</span>
                </div>
            </div>
        );
    }

    return (
        <div className="chart-card light animate-fade-in word-cloud-card">
            <div className="chart-card-header">
                <h3>Problem Keywords</h3>
                <span className="word-count-badge">{words.length} keywords</span>
            </div>
            <div className="word-cloud-container">
                {words.length === 0 ? (
                    <div className="word-cloud-empty">
                        <p>No keywords found</p>
                    </div>
                ) : (
                    <div className="word-cloud-bubble">
                        {words.map((word, index) => (
                            <div
                                key={word.word}
                                className={`word-bubble ${getSizeClass(word.count, maxCount)}`}
                                style={{
                                    background: getGradientStyle(index),
                                    animationDelay: `${index * 0.03}s`,
                                }}
                                title={`${word.word}: ${word.count} occurrences`}
                            >
                                <span className="word-text">{word.word}</span>
                                <span className="word-count">{word.count}</span>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default WordCloud;
