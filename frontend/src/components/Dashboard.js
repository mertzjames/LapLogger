import React, { useState, useEffect } from 'react';
import { swimmersAPI, timesAPI } from '../services/api';

function Dashboard() {
  const [stats, setStats] = useState({
    totalSwimmers: 0,
    totalTimes: 0,
    recentTimes: [],
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [swimmersResponse, timesResponse] = await Promise.all([
          swimmersAPI.getAll(),
          timesAPI.getAll(),
        ]);

        const swimmers = swimmersResponse.data;
        const times = timesResponse.data;

        setStats({
          totalSwimmers: swimmers.length,
          totalTimes: times.length,
          recentTimes: times.slice(0, 5),
        });
      } catch (error) {
        console.error('Error fetching dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return <div className="card">Loading dashboard...</div>;
  }

  return (
    <div>
      <h1>Dashboard</h1>
      
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-number">{stats.totalSwimmers}</div>
          <div className="stat-label">Total Swimmers</div>
        </div>
        
        <div className="stat-card">
          <div className="stat-number">{stats.totalTimes}</div>
          <div className="stat-label">Total Times Logged</div>
        </div>
        
        <div className="stat-card">
          <div className="stat-number">
            {stats.recentTimes.length > 0 ? stats.recentTimes[0].formatted_time : 'N/A'}
          </div>
          <div className="stat-label">Latest Time</div>
        </div>
      </div>

      <div className="card">
        <h2>Recent Times</h2>
        {stats.recentTimes.length === 0 ? (
          <p>No times logged yet. <a href="/add-time">Log your first time!</a></p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Swimmer</th>
                <th>Event</th>
                <th>Time</th>
                <th>Date</th>
                <th>Meet</th>
              </tr>
            </thead>
            <tbody>
              {stats.recentTimes.map((time) => (
                <tr key={time.id}>
                  <td>{time.swimmer_name}</td>
                  <td>{time.event_name}</td>
                  <td><strong>{time.formatted_time}</strong></td>
                  <td>{new Date(time.recorded_at).toLocaleDateString()}</td>
                  <td>{time.meet_name || 'Practice'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
