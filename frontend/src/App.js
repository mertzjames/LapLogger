import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import './App.css';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import PrivateRoute from './components/PrivateRoute';
import Login from './components/Login';
import Register from './components/Register';
import Dashboard from './components/Dashboard';
import Swimmers from './components/Swimmers';
import Times from './components/Times';
import AddTime from './components/AddTime';

function AppContent() {
  const { user, logout, isAuthenticated } = useAuth();

  return (
    <div className="App">
      <nav className="navbar">
        <div className="nav-container">
          <h1 className="nav-title">LapLogger</h1>
          {isAuthenticated && (
            <ul className="nav-menu">
              <li className="nav-item">
                <Link to="/" className="nav-link">Dashboard</Link>
              </li>
              <li className="nav-item">
                <Link to="/swimmers" className="nav-link">Swimmers</Link>
              </li>
              <li className="nav-item">
                <Link to="/times" className="nav-link">Times</Link>
              </li>
              <li className="nav-item">
                <Link to="/add-time" className="nav-link">Log Time</Link>
              </li>
            </ul>
          )}
          {isAuthenticated && (
            <div className="nav-user">
              <span className="nav-username">Welcome, {user?.username}</span>
              <button onClick={logout} className="nav-logout">Logout</button>
            </div>
          )}
        </div>
      </nav>

      <main className="main-content">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/" element={
            <PrivateRoute>
              <Dashboard />
            </PrivateRoute>
          } />
          <Route path="/swimmers" element={
            <PrivateRoute>
              <Swimmers />
            </PrivateRoute>
          } />
          <Route path="/times" element={
            <PrivateRoute>
              <Times />
            </PrivateRoute>
          } />
          <Route path="/add-time" element={
            <PrivateRoute>
              <AddTime />
            </PrivateRoute>
          } />
        </Routes>
      </main>
    </div>
  );
}

function App() {
  return (
    <Router>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </Router>
  );
}

export default App;
