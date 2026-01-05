import { createContext, useContext, useState, useEffect } from 'react';
import authService from '../services/authApi';

const AuthContext = createContext(null);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Check for existing session
        const accessToken = localStorage.getItem('access_token');
        const userData = localStorage.getItem('user_data');

        if (accessToken && userData) {
            try {
                setUser(JSON.parse(userData));
            } catch (e) {
                localStorage.removeItem('access_token');
                localStorage.removeItem('refresh_token');
                localStorage.removeItem('user_data');
            }
        }
        setLoading(false);
    }, []);

    const login = async (email, password) => {
        try {
            const response = await authService.login(email, password);

            if (response.success && response.data) {
                const { access_token, refresh_token, user: userData } = response.data;

                localStorage.setItem('access_token', access_token);
                localStorage.setItem('refresh_token', refresh_token);
                localStorage.setItem('user_data', JSON.stringify(userData));

                setUser(userData);
                return { success: true };
            }

            return { success: false, message: response.message || 'Login failed' };
        } catch (error) {
            const message = error.response?.data?.message || 'Login failed. Please try again.';
            return { success: false, message };
        }
    };

    const register = async (userData) => {
        try {
            const response = await authService.register(userData);
            return { success: response.success, message: response.message };
        } catch (error) {
            const message = error.response?.data?.message || 'Registration failed. Please try again.';
            return { success: false, message };
        }
    };

    const forgotPassword = async (email) => {
        try {
            const response = await authService.forgotPassword(email);
            return { success: response.success, message: response.message };
        } catch (error) {
            const message = error.response?.data?.message || 'Request failed. Please try again.';
            return { success: false, message };
        }
    };

    const logout = async () => {
        const accessToken = localStorage.getItem('access_token');
        const refreshToken = localStorage.getItem('refresh_token');

        await authService.logout(accessToken, refreshToken);

        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_data');

        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, loading, login, register, forgotPassword, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export default AuthContext;
