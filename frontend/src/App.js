import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import './App.css';
import Dashboard from './components/Dashboard';
import Swimmers from './components/Swimmers';
import Times from './components/Times';
import AddTime from './components/AddTime';

function App() {
  return (
    <Router>
      <div className="App">
        <nav className="navbar">
          <div className="nav-container">
            <h1 className="nav-title">LapLogger</h1>
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
          </div>
        </nav>

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/swimmers" element={<Swimmers />} />
            <Route path="/times" element={<Times />} />
            <Route path="/add-time" element={<AddTime />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
