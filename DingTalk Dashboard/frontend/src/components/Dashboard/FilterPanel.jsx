import { useState, useEffect } from 'react';
import './FilterPanel.css';

const FilterPanel = ({
    filters,
    onFilterChange,
    filterOptions,
    loading
}) => {
    const [isExpanded, setIsExpanded] = useState(true);
    const [datePreset, setDatePreset] = useState('');

    // Date preset options
    const datePresets = [
        { label: 'Today', value: 'today' },
        { label: 'Yesterday', value: 'yesterday' },
        { label: 'This Week', value: 'this_week' },
        { label: 'This Month', value: 'this_month' },
        { label: 'Custom', value: 'custom' },
    ];

    const handleDatePreset = (preset) => {
        setDatePreset(preset);
        const today = new Date();
        let startDate = null;
        let endDate = null;

        switch (preset) {
            case 'today':
                startDate = today.toISOString().split('T')[0];
                endDate = today.toISOString().split('T')[0];
                break;
            case 'yesterday':
                const yesterday = new Date(today);
                yesterday.setDate(yesterday.getDate() - 1);
                startDate = yesterday.toISOString().split('T')[0];
                endDate = yesterday.toISOString().split('T')[0];
                break;
            case 'this_week':
                const startOfWeek = new Date(today);
                startOfWeek.setDate(today.getDate() - today.getDay());
                startDate = startOfWeek.toISOString().split('T')[0];
                endDate = today.toISOString().split('T')[0];
                break;
            case 'this_month':
                const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
                startDate = startOfMonth.toISOString().split('T')[0];
                endDate = today.toISOString().split('T')[0];
                break;
            case 'custom':
            default:
                // Don't change dates for custom
                return;
        }

        onFilterChange('start_date', startDate);
        onFilterChange('end_date', endDate);
    };

    // Get active filter count
    const activeFilterCount = [
        filters.search,
        filters.business_id,
        filters.status,
        filters.department,
        filters.ditujukan_kepada,
        filters.dilaporkan_oleh,
        filters.kategori,
        filters.start_date,
        filters.end_date,
    ].filter(Boolean).length;

    // Clear all filters
    const clearAllFilters = () => {
        onFilterChange('search', '');
        onFilterChange('business_id', '');
        onFilterChange('status', '');
        onFilterChange('department', '');
        onFilterChange('ditujukan_kepada', '');
        onFilterChange('dilaporkan_oleh', '');
        onFilterChange('kategori', '');
        onFilterChange('start_date', '');
        onFilterChange('end_date', '');
        setDatePreset('');
    };

    // Filter chip component
    const FilterChip = ({ label, value, onRemove }) => (
        <div className="filter-chip">
            <span className="filter-chip-label">{label}:</span>
            <span className="filter-chip-value">{value}</span>
            <button className="filter-chip-remove" onClick={onRemove}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
            </button>
        </div>
    );

    return (
        <div className="filter-panel">
            <div className="filter-panel-header">
                <button
                    className="filter-toggle-btn"
                    onClick={() => setIsExpanded(!isExpanded)}
                >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6" strokeLinecap="round" />
                    </svg>
                    Advanced Filters
                    {activeFilterCount > 0 && (
                        <span className="filter-count-badge">{activeFilterCount}</span>
                    )}
                </button>
                {activeFilterCount > 0 && (
                    <button className="clear-filters-btn" onClick={clearAllFilters}>
                        Clear All
                    </button>
                )}
                <svg
                    className={`filter-expand-icon ${isExpanded ? 'expanded' : ''}`}
                    width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
                >
                    <path d="M6 9l6 6 6-6" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
            </div>

            {/* Active Filter Chips */}
            {activeFilterCount > 0 && (
                <div className="filter-chips">
                    {filters.search && (
                        <FilterChip label="Search" value={filters.search} onRemove={() => onFilterChange('search', '')} />
                    )}
                    {filters.business_id && (
                        <FilterChip label="Business ID" value={filters.business_id} onRemove={() => onFilterChange('business_id', '')} />
                    )}
                    {filters.status && (
                        <FilterChip label="Status" value={filters.status} onRemove={() => onFilterChange('status', '')} />
                    )}
                    {filters.department && (
                        <FilterChip label="Department" value={filters.department} onRemove={() => onFilterChange('department', '')} />
                    )}
                    {filters.ditujukan_kepada && (
                        <FilterChip label="Ditujukan Kepada" value={filters.ditujukan_kepada} onRemove={() => onFilterChange('ditujukan_kepada', '')} />
                    )}
                    {filters.dilaporkan_oleh && (
                        <FilterChip label="Dilaporkan Oleh" value={filters.dilaporkan_oleh} onRemove={() => onFilterChange('dilaporkan_oleh', '')} />
                    )}
                    {filters.kategori && (
                        <FilterChip label="Kategori" value={filters.kategori} onRemove={() => onFilterChange('kategori', '')} />
                    )}
                    {(filters.start_date || filters.end_date) && (
                        <FilterChip
                            label="Date Range"
                            value={`${filters.start_date || '...'} - ${filters.end_date || '...'}`}
                            onRemove={() => {
                                onFilterChange('start_date', '');
                                onFilterChange('end_date', '');
                                setDatePreset('');
                            }}
                        />
                    )}
                </div>
            )}

            <div className={`filter-panel-content ${isExpanded ? 'expanded' : ''}`}>
                {/* Row 1: Search and Business ID */}
                <div className="filter-row">
                    <div className="filter-group">
                        <label>Search</label>
                        <div className="input-with-icon">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <circle cx="11" cy="11" r="8" />
                                <path d="M21 21l-4.35-4.35" />
                            </svg>
                            <input
                                type="text"
                                placeholder="Search by title, name, project..."
                                value={filters.search || ''}
                                onChange={(e) => onFilterChange('search', e.target.value)}
                            />
                        </div>
                    </div>
                    <div className="filter-group">
                        <label>Business ID</label>
                        <input
                            type="text"
                            placeholder="Enter Business ID"
                            value={filters.business_id || ''}
                            onChange={(e) => onFilterChange('business_id', e.target.value)}
                        />
                    </div>
                    <div className="filter-group">
                        <label>Status</label>
                        <select
                            value={filters.status || ''}
                            onChange={(e) => onFilterChange('status', e.target.value)}
                        >
                            <option value="">All Status</option>
                            {filterOptions?.statuses?.map(status => (
                                <option key={status} value={status}>{status}</option>
                            ))}
                        </select>
                    </div>
                </div>

                {/* Row 2: Date Range */}
                <div className="filter-row date-row">
                    <div className="filter-group date-presets">
                        <label>Date Range</label>
                        <div className="preset-buttons">
                            {datePresets.map(preset => (
                                <button
                                    key={preset.value}
                                    className={`preset-btn ${datePreset === preset.value ? 'active' : ''}`}
                                    onClick={() => handleDatePreset(preset.value)}
                                >
                                    {preset.label}
                                </button>
                            ))}
                        </div>
                    </div>
                    <div className="filter-group">
                        <label>Start Date</label>
                        <input
                            type="date"
                            value={filters.start_date || ''}
                            onChange={(e) => {
                                onFilterChange('start_date', e.target.value);
                                setDatePreset('custom');
                            }}
                        />
                    </div>
                    <div className="filter-group">
                        <label>End Date</label>
                        <input
                            type="date"
                            value={filters.end_date || ''}
                            onChange={(e) => {
                                onFilterChange('end_date', e.target.value);
                                setDatePreset('custom');
                            }}
                        />
                    </div>
                </div>

                {/* Row 3: Department and Category Filters */}
                <div className="filter-row">
                    <div className="filter-group">
                        <label>Department</label>
                        <select
                            value={filters.department || ''}
                            onChange={(e) => onFilterChange('department', e.target.value)}
                        >
                            <option value="">All Departments</option>
                            {filterOptions?.departments?.map(dept => (
                                <option key={dept} value={dept}>{dept}</option>
                            ))}
                        </select>
                    </div>
                    <div className="filter-group">
                        <label>Ditujukan Kepada</label>
                        <select
                            value={filters.ditujukan_kepada || ''}
                            onChange={(e) => onFilterChange('ditujukan_kepada', e.target.value)}
                        >
                            <option value="">All</option>
                            {filterOptions?.ditujukan_kepada?.map(item => (
                                <option key={item} value={item}>{item}</option>
                            ))}
                        </select>
                    </div>
                    <div className="filter-group">
                        <label>Dilaporkan Oleh</label>
                        <select
                            value={filters.dilaporkan_oleh || ''}
                            onChange={(e) => onFilterChange('dilaporkan_oleh', e.target.value)}
                        >
                            <option value="">All</option>
                            {filterOptions?.dilaporkan_oleh?.map(item => (
                                <option key={item} value={item}>{item}</option>
                            ))}
                        </select>
                    </div>
                    <div className="filter-group">
                        <label>Kategori</label>
                        <select
                            value={filters.kategori || ''}
                            onChange={(e) => onFilterChange('kategori', e.target.value)}
                        >
                            <option value="">All Categories</option>
                            {filterOptions?.kategori?.map(kat => (
                                <option key={kat} value={kat}>{kat}</option>
                            ))}
                        </select>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default FilterPanel;
