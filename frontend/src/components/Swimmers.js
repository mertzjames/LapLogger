import React, { useState, useEffect } from 'react';
import { swimmersAPI } from '../services/api';

function Swimmers() {
  const [swimmers, setSwimmers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
  });

  useEffect(() => {
    fetchSwimmers();
  }, []);

  const fetchSwimmers = async () => {
    try {
      const response = await swimmersAPI.getAll();
      setSwimmers(response.data);
    } catch (error) {
      console.error('Error fetching swimmers:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await swimmersAPI.create(formData);
      setFormData({ name: '', email: '' });
      setShowForm(false);
      fetchSwimmers();
    } catch (error) {
      console.error('Error creating swimmer:', error);
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

  return (
    <div>
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1>Swimmers</h1>
          <button 
            className="btn btn-primary"
            onClick={() => setShowForm(!showForm)}
          >
            {showForm ? 'Cancel' : 'Add Swimmer'}
          </button>
        </div>

        {showForm && (
          <form onSubmit={handleSubmit} style={{ marginTop: '1rem' }}>
            <div className="form-group">
              <label>Name</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
              />
            </div>
            <div className="form-group">
              <label>Email (optional)</label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
              />
            </div>
            <button type="submit" className="btn btn-success">
              Add Swimmer
            </button>
          </form>
        )}
      </div>

      <div className="card">
        <h2>All Swimmers ({swimmers.length})</h2>
        {swimmers.length === 0 ? (
          <p>No swimmers added yet.</p>
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
