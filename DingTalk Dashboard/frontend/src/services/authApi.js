import axios from 'axios';

// Use backend proxy to avoid CORS issues with external auth API
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8087';

// Auth service - requests go through backend which proxies to external auth API
const authApi = axios.create({
    baseURL: API_BASE,
    headers: {
        'Content-Type': 'application/json',
    },
});

export const authService = {
    // Login via backend proxy
    login: async (email, password, frontendId = 'dingtalk-dashboard') => {
        const response = await authApi.post('/api/v1/auth/login', {
            email,
            password,
            frontend_id: frontendId,
            device_info: {
                browser: navigator.userAgent,
                platform: navigator.platform,
            },
        });
        return response.data;
    },

    // Register new user via backend proxy
    register: async (userData) => {
        const response = await authApi.post('/api/v1/auth/register', userData);
        return response.data;
    },

    // Forgot password via backend proxy
    forgotPassword: async (email) => {
        const response = await authApi.post('/api/v1/auth/forgot-password', { email });
        return response.data;
    },

    // Reset password via backend proxy
    resetPassword: async (token, newPassword) => {
        const response = await authApi.post('/api/v1/auth/reset-password', {
            token,
            new_password: newPassword,
        });
        return response.data;
    },

    // Refresh access token via backend proxy
    refreshToken: async (refreshToken) => {
        const response = await authApi.post('/api/v1/auth/refresh', {
            refresh_token: refreshToken,
        });
        return response.data;
    },

    // Logout via backend proxy
    logout: async (accessToken, refreshToken) => {
        try {
            await authApi.post(
                '/api/v1/auth/logout',
                { refresh_token: refreshToken },
                { headers: { Authorization: `Bearer ${accessToken}` } }
            );
        } catch (error) {
            // Ignore logout errors
            console.error('Logout error:', error);
        }
    },
};

export default authService;
