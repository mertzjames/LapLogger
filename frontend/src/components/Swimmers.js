import React, { useState, useEffect } from 'react';
import { swimmersAPI } from '../services/api';

function Swimmers() {
  const [swimmers, setSwimmers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
  });

  useEffect(() => {
    fetchSwimmers();
  }, []);

  const fetchSwimmers = async () => {
    try {
      setError(null);
      const response = await swimmersAPI.getAll();
      setSwimmers(response.data || []);
    } catch (error) {
      console.error('Error fetching swimmers:', error);
      setError('Failed to load swimmers. Please check if the backend server is running.');
      setSwimmers([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await swimmersAPI.create(formData);
      setFormData({ name: '', email: '' });
      setShowForm(false);
      fetchSwimmers();
      alert('Swimmer added successfully!');
    } catch (error) {
      console.error('Error creating swimmer:', error);
      alert('Failed to add swimmer. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleInputChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  if (loading) {
    return <div className="card">Loading swimmers...</div>;
  }

  if (error) {
    return (
      <div className="card">
        <h1>Swimmers</h1>
        <div style={{ color: '#e74c3c', marginBottom: '1rem' }}>
          <strong>Error:</strong> {error}
        </div>
        <button 
          className="btn btn-primary" 
          onClick={fetchSwimmers}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1>Swimmers</h1>
          <button 
            className="btn btn-primary"
            onClick={() => setShowForm(!showForm)}
            disabled={submitting}
          >
            {showForm ? 'Cancel' : 'Add Swimmer'}
          </button>
        </div>

        {showForm && (
          <form onSubmit={handleSubmit} style={{ marginTop: '1rem' }}>
            <div className="form-group">
              <label>Name *</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
                placeholder="Enter swimmer's name"
              />
            </div>
            <div className="form-group">
              <label>Email (optional)</label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                placeholder="Enter swimmer's email"
              />
            </div>
            <button 
              type="submit" 
              className="btn btn-success"
              disabled={submitting}
            >
              {submitting ? 'Adding...' : 'Add Swimmer'}
            </button>
          </form>
        )}
      </div>

      <div className="card">
        <h2>All Swimmers ({swimmers.length})</h2>
        {swimmers.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <p style={{ fontSize: '1.1rem', color: '#7f8c8d', marginBottom: '1rem' }}>
              No swimmers have been added yet.
            </p>
            <p style={{ marginBottom: '1rem' }}>
              Add your first swimmer to start tracking swim times!
            </p>
            <button 
              className="btn btn-primary"
              onClick={() => setShowForm(true)}
            >
              Add First Swimmer
            </button>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Added</th>
              </tr>
            </thead>
            <tbody>
              {swimmers.map((swimmer) => (
                <tr key={swimmer.id}>
                  <td><strong>{swimmer.name}</strong></td>
                  <td>{swimmer.email || 'Not provided'}</td>
                  <td>{new Date(swimmer.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default Swimmers;
