import React, { useState, useEffect } from 'react';
import { timesAPI } from '../services/api';

function Times() {
  const [times, setTimes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    fetchTimes();
  }, []);

  const fetchTimes = async () => {
    try {
      setError(null);
      const response = await timesAPI.getAll();
      setTimes(response.data || []);
    } catch (error) {
      console.error('Error fetching times:', error);
      setError('Failed to load times. Please check if the backend server is running.');
      setTimes([]);
    } finally {
      setLoading(false);
    }
  };

  const filteredTimes = times.filter(time => {
    if (filter === 'all') return true;
    if (filter === 'practice') return !time.meet_name;
    if (filter === 'meets') return time.meet_name;
    return true;
  });

  if (loading) {
    return <div className="card">Loading times...</div>;
  }

  if (error) {
    return (
      <div className="card">
        <h1>Times</h1>
        <div style={{ color: '#e74c3c', marginBottom: '1rem' }}>
          <strong>Error:</strong> {error}
        </div>
        <button 
          className="btn btn-primary" 
          onClick={fetchTimes}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="card">
        <h1>All Times</h1>
        
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ marginRight: '1rem' }}>Filter:</label>
          <select 
            value={filter} 
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '0.5rem' }}
          >
            <option value="all">All Times</option>
            <option value="practice">Practice Only</option>
            <option value="meets">Meet Times Only</option>
          </select>
        </div>

        {times.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <p style={{ fontSize: '1.1rem', color: '#7f8c8d', marginBottom: '1rem' }}>
              No swim times have been logged yet.
            </p>
            <p style={{ marginBottom: '1rem' }}>
              Start tracking performance by logging your first swim time!
            </p>
            <a href="/add-time" className="btn btn-primary">
              Log First Time
            </a>
          </div>
        ) : filteredTimes.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <p style={{ fontSize: '1.1rem', color: '#7f8c8d', marginBottom: '1rem' }}>
              No times match the selected filter.
            </p>
            <p style={{ marginBottom: '1rem' }}>
              Try changing the filter or <a href="/add-time">log a new time</a>.
            </p>
          </div>
        ) : (
          <div>
            <p>Showing {filteredTimes.length} of {times.length} times</p>
            <table className="table">
              <thead>
                <tr>
                  <th>Swimmer</th>
                  <th>Event</th>
                  <th>Time</th>
                  <th>Meet</th>
                  <th>Date</th>
                  <th>Notes</th>
                </tr>
              </thead>
              <tbody>
                {filteredTimes.map((time) => (
                  <tr key={time.id}>
                    <td><strong>{time.swimmer_name}</strong></td>
                    <td>{time.event_name}</td>
                    <td>
                      <strong style={{ fontSize: '1.1em', color: '#2c3e50' }}>
                        {time.formatted_time}
                      </strong>
                    </td>
                    <td>
                      {time.meet_name ? (
                        <span style={{ color: '#27ae60' }}>{time.meet_name}</span>
                      ) : (
                        <span style={{ color: '#7f8c8d' }}>Practice</span>
                      )}
                    </td>
                    <td>{new Date(time.recorded_at).toLocaleDateString()}</td>
                    <td>{time.notes || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

export default Times;
