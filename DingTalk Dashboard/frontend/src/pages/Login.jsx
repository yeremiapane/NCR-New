import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import './Login.css';

const Login = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        const result = await login(email, password);

        if (result.success) {
            navigate('/');
        } else {
            setError(result.message);
        }

        setLoading(false);
    };

    return (
        <div className="login-container">
            <div className="login-background">
                <div className="login-orb login-orb-1"></div>
                <div className="login-orb login-orb-2"></div>
                <div className="login-orb login-orb-3"></div>
            </div>

            <div className="login-card">
                <div className="login-header">
                    <div className="login-logo">
                        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                            <rect width="40" height="40" rx="10" fill="url(#logo-gradient)" />
                            <path d="M12 20L18 26L28 14" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
                            <defs>
                                <linearGradient id="logo-gradient" x1="0" y1="0" x2="40" y2="40">
                                    <stop stopColor="#6366f1" />
                                    <stop offset="1" stopColor="#818cf8" />
                                </linearGradient>
                            </defs>
                        </svg>
                    </div>
                    <h1>DingTalk Dashboard</h1>
                    <p className="text-secondary">Sign in to access approval data</p>
                </div>

                <form onSubmit={handleSubmit} className="login-form">
                    {error && (
                        <div className="login-error animate-fade-in">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm-1-7v2h2v-2h-2zm0-8v6h2V7h-2z" />
                            </svg>
                            {error}
                        </div>
                    )}

                    <div className="input-group">
                        <label htmlFor="email">Email</label>
                        <input
                            id="email"
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="Enter your email"
                            required
                            autoComplete="email"
                        />
                    </div>

                    <div className="input-group">
                        <label htmlFor="password">Password</label>
                        <input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Enter your password"
                            required
                            autoComplete="current-password"
                        />
                    </div>

                    <button type="submit" className="btn btn-primary btn-lg w-full" disabled={loading}>
                        {loading ? (
                            <>
                                <span className="spinner"></span>
                                Signing in...
                            </>
                        ) : (
                            'Sign In'
                        )}
                    </button>
                </form>

                <div className="login-footer">
                    <Link to="/forgot-password" className="login-link">Forgot password?</Link>
                    <span className="text-muted">•</span>
                    <Link to="/register" className="login-link">Create account</Link>
                </div>
            </div>
        </div>
    );
};

export default Login;
