import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import PrivateRoute from './components/PrivateRoute';
import Layout from './components/Layout';
import Login from './components/Login';
import LeagueSelect from './components/LeagueSelect';
import Dashboard from './components/Dashboard';
import Teams from './components/Teams';
import Swimmers from './components/Swimmers';
import Meets from './components/Meets';
import Events from './components/Events';
import TimeEntry from './components/TimeEntry';
import PublicResults from './components/PublicResults';
import './App.css';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/results/:leagueSlug/meets/:meetId"
            element={<PublicResults />}
          />
          <Route
            element={
              <PrivateRoute>
                <Layout />
              </PrivateRoute>
            }
          >
            <Route index element={<LeagueSelect />} />
            <Route path="leagues/:leagueId" element={<Dashboard />} />
            <Route path="leagues/:leagueId/teams" element={<Teams />} />
            <Route path="leagues/:leagueId/swimmers" element={<Swimmers />} />
            <Route path="leagues/:leagueId/meets" element={<Meets />} />
            <Route
              path="leagues/:leagueId/meets/:meetId/events"
              element={<Events />}
            />
            <Route
              path="leagues/:leagueId/meets/:meetId/events/:eventId/times"
              element={<TimeEntry />}
            />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
