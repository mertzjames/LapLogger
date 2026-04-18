import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import type { Meet } from '../types';
import { listMeets, createMeet, updateMeet, deleteMeet } from '../services/api';

const emptyForm = {
  name: '',
  location: '',
  meet_date: '',
  is_public: false,
};

export default function Meets() {
  const { leagueId } = useParams<{ leagueId: string }>();
  const [meets, setMeets] = useState<Meet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);

  const load = () => {
    if (!leagueId) return;
    listMeets(leagueId)
      .then(setMeets)
      .catch(() => setError('Failed to load meets'))
      .finally(() => setLoading(false));
  };

  useEffect(load, [leagueId]);

  const resetForm = () => {
    setForm(emptyForm);
    setEditingId(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leagueId) return;
    setError('');
    try {
      if (editingId) {
        await updateMeet(leagueId, editingId, form);
      } else {
        await createMeet(leagueId, form);
      }
      resetForm();
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save meet');
    }
  };

  const handleEdit = (m: Meet) => {
    setEditingId(m.id);
    setForm({
      name: m.name,
      location: m.location || '',
      meet_date: m.meet_date,
      is_public: m.is_public,
    });
  };

  const handleDelete = async (id: string) => {
    if (!leagueId) return;
    setError('');
    try {
      await deleteMeet(leagueId, id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete meet');
    }
  };

  const togglePublic = async (meet: Meet) => {
    if (!leagueId) return;
    try {
      await updateMeet(leagueId, meet.id, { is_public: !meet.is_public });
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update');
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="meets-page">
      <h1>Meets</h1>
      {error && <div className="error-msg">{error}</div>}

      <form onSubmit={handleSubmit} className="crud-form">
        <input
          type="text"
          placeholder="Meet name"
          value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Location"
          value={form.location}
          onChange={(e) => setForm({ ...form, location: e.target.value })}
        />
        <input
          type="date"
          value={form.meet_date}
          onChange={(e) => setForm({ ...form, meet_date: e.target.value })}
          required
        />
        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={form.is_public}
            onChange={(e) => setForm({ ...form, is_public: e.target.checked })}
          />
          Public
        </label>
        <div className="form-actions">
          <button type="submit" className="btn btn-primary">
            {editingId ? 'Update' : 'Add Meet'}
          </button>
          {editingId && (
            <button type="button" className="btn" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      {meets.length === 0 ? (
        <p className="empty-state">No meets yet.</p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Date</th>
              <th>Location</th>
              <th>Public</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {meets.map((meet) => (
              <tr key={meet.id}>
                <td>
                  <Link to={`/leagues/${leagueId}/meets/${meet.id}/events`}>
                    {meet.name}
                  </Link>
                </td>
                <td>{meet.meet_date}</td>
                <td>{meet.location || '—'}</td>
                <td>
                  <button
                    className={`btn btn-sm ${meet.is_public ? 'btn-success' : ''}`}
                    onClick={() => togglePublic(meet)}
                  >
                    {meet.is_public ? 'Yes' : 'No'}
                  </button>
                </td>
                <td className="actions">
                  <button className="btn btn-sm" onClick={() => handleEdit(meet)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => handleDelete(meet.id)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
