import { Outlet, Link, useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { leagueId } = useParams();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="app-layout">
      <nav className="navbar">
        <Link to="/" className="nav-brand">LapLogger</Link>
        <div className="nav-links">
          {leagueId && (
            <>
              <Link to={`/leagues/${leagueId}`}>Dashboard</Link>
              <Link to={`/leagues/${leagueId}/teams`}>Teams</Link>
              <Link to={`/leagues/${leagueId}/swimmers`}>Swimmers</Link>
              <Link to={`/leagues/${leagueId}/meets`}>Meets</Link>
            </>
          )}
        </div>
        <div className="nav-user">
          <span>{user?.name}</span>
          <button onClick={handleLogout} className="btn btn-sm">Logout</button>
        </div>
      </nav>
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}
