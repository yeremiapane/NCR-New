import axios from 'axios';

// Dashboard API service for approval data
const dashboardApi = axios.create({
    baseURL: '/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
});

// Add auth token to requests
dashboardApi.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Handle token refresh on 401
dashboardApi.interceptors.response.use(
    (response) => response,
    async (error) => {
        if (error.response?.status === 401) {
            // Token expired, could implement refresh here
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export const dashboardService = {
    // Get approval list
    getApprovals: async (params = {}) => {
        const response = await dashboardApi.get('/approvals', { params });
        return response.data;
    },

    // Get single approval with details
    getApproval: async (id) => {
        const response = await dashboardApi.get(`/approvals/${id}`);
        return response.data;
    },

    // Get dashboard statistics
    getStats: async (filters = {}) => {
        const response = await dashboardApi.get('/approvals/stats', { params: filters });
        return response.data;
    },

    // Get filter options (distinct values for dropdowns)
    getFilterOptions: async () => {
        const response = await dashboardApi.get('/approvals/filter-options');
        return response.data;
    },

    // Get sync logs
    getSyncLogs: async (params = {}) => {
        const response = await dashboardApi.get('/sync/logs', { params });
        return response.data;
    },

    // Trigger manual sync
    triggerSync: async () => {
        const response = await dashboardApi.post('/sync/trigger');
        return response.data;
    },
};

export default dashboardService;
