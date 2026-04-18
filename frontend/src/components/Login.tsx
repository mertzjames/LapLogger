import { useAuth } from '../contexts/AuthContext';
import { Navigate } from 'react-router-dom';

export default function Login() {
  const { user, loading, login } = useAuth();

  if (loading) return <div className="loading">Loading...</div>;
  if (user) return <Navigate to="/" replace />;

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>LapLogger</h1>
        <p>Multi-tenant swim time tracking for leagues and teams.</p>
        <button onClick={login} className="btn btn-primary btn-google">
          Sign in with Google
        </button>
      </div>
    </div>
  );
}
