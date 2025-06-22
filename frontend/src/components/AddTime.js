import React, { useState, useEffect } from 'react';
import { timesAPI, swimmersAPI, eventsAPI } from '../services/api';

function AddTime() {
  const [swimmers, setSwimmers] = useState([]);
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    swimmer_id: '',
    event_id: '',
    minutes: '',
    seconds: '',
    milliseconds: '',
    notes: '',
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [swimmersResponse, eventsResponse] = await Promise.all([
        swimmersAPI.getAll(),
        eventsAPI.getAll(),
      ]);
      setSwimmers(swimmersResponse.data);
      setEvents(eventsResponse.data);
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      // Convert time to milliseconds
      const minutes = parseInt(formData.minutes) || 0;
      const seconds = parseInt(formData.seconds) || 0;
      const milliseconds = parseInt(formData.milliseconds) || 0;
      
      const totalMs = (minutes * 60 * 1000) + (seconds * 1000) + milliseconds;

      const timeData = {
        swimmer_id: parseInt(formData.swimmer_id),
        event_id: parseInt(formData.event_id),
        time_ms: totalMs,
        notes: formData.notes,
      };

      await timesAPI.create(timeData);
      
      // Reset form
      setFormData({
        swimmer_id: '',
        event_id: '',
        minutes: '',
        seconds: '',
        milliseconds: '',
        notes: '',
      });

      alert('Time logged successfully!');
    } catch (error) {
      console.error('Error logging time:', error);
      alert('Error logging time. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <div className="card">Loading...</div>;
  }

  const formatTime = () => {
    const minutes = parseInt(formData.minutes) || 0;
    const seconds = parseInt(formData.seconds) || 0;
    const milliseconds = parseInt(formData.milliseconds) || 0;
    return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}.${milliseconds.toString().padStart(3, '0')}`;
  };

  return (
    <div>
      <div className="card">
        <h1>Log New Time</h1>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Swimmer</label>
            <select
              name="swimmer_id"
              value={formData.swimmer_id}
              onChange={handleInputChange}
              required
            >
              <option value="">Select a swimmer</option>
              {swimmers.map((swimmer) => (
                <option key={swimmer.id} value={swimmer.id}>
                  {swimmer.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Event</label>
            <select
              name="event_id"
              value={formData.event_id}
              onChange={handleInputChange}
              required
            >
              <option value="">Select an event</option>
              {events.map((event) => (
                <option key={event.id} value={event.id}>
                  {event.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Time</label>
            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              <input
                type="number"
                name="minutes"
                value={formData.minutes}
                onChange={handleInputChange}
                placeholder="MM"
                min="0"
                max="99"
                style={{ width: '80px' }}
              />
              <span>:</span>
              <input
                type="number"
                name="seconds"
                value={formData.seconds}
                onChange={handleInputChange}
                placeholder="SS"
                min="0"
                max="59"
                style={{ width: '80px' }}
              />
              <span>.</span>
              <input
                type="number"
                name="milliseconds"
                value={formData.milliseconds}
                onChange={handleInputChange}
                placeholder="MMM"
                min="0"
                max="999"
                style={{ width: '100px' }}
              />
            </div>
            {(formData.minutes || formData.seconds || formData.milliseconds) && (
              <small style={{ color: '#7f8c8d', marginTop: '0.5rem', display: 'block' }}>
                Preview: {formatTime()}
              </small>
            )}
          </div>

          <div className="form-group">
            <label>Notes (optional)</label>
            <textarea
              name="notes"
              value={formData.notes}
              onChange={handleInputChange}
              placeholder="Any additional notes about this time..."
              rows="3"
            />
          </div>

          <button 
            type="submit" 
            className="btn btn-success"
            disabled={submitting}
          >
            {submitting ? 'Logging Time...' : 'Log Time'}
          </button>
        </form>
      </div>

      {swimmers.length === 0 && (
        <div className="card">
          <h3>No Swimmers Found</h3>
          <p>You need to add at least one swimmer before logging times.</p>
          <a href="/swimmers" className="btn btn-primary">Add Swimmers</a>
        </div>
      )}
    </div>
  );
}

export default AddTime;
