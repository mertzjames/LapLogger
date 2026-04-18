import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { Team } from '../types';
import { listTeams, createTeam, updateTeam, deleteTeam } from '../services/api';

export default function Teams() {
  const { leagueId } = useParams<{ leagueId: string }>();
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ name: '', short_name: '' });

  const load = () => {
    if (!leagueId) return;
    listTeams(leagueId)
      .then(setTeams)
      .catch(() => setError('Failed to load teams'))
      .finally(() => setLoading(false));
  };

  useEffect(load, [leagueId]);

  const resetForm = () => {
    setForm({ name: '', short_name: '' });
    setEditingId(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leagueId || !form.name.trim()) return;
    setError('');
    try {
      if (editingId) {
        await updateTeam(leagueId, editingId, form);
      } else {
        await createTeam(leagueId, form);
      }
      resetForm();
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save team');
    }
  };

  const handleEdit = (team: Team) => {
    setEditingId(team.id);
    setForm({ name: team.name, short_name: team.short_name || '' });
  };

  const handleDelete = async (teamId: string) => {
    if (!leagueId) return;
    setError('');
    try {
      await deleteTeam(leagueId, teamId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete team');
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="teams-page">
      <h1>Teams</h1>
      {error && <div className="error-msg">{error}</div>}

      <form onSubmit={handleSubmit} className="crud-form">
        <input
          type="text"
          placeholder="Team name"
          value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Short name (optional)"
          value={form.short_name}
          onChange={(e) => setForm({ ...form, short_name: e.target.value })}
        />
        <div className="form-actions">
          <button type="submit" className="btn btn-primary">
            {editingId ? 'Update' : 'Add Team'}
          </button>
          {editingId && (
            <button type="button" className="btn" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      {teams.length === 0 ? (
        <p className="empty-state">No teams yet.</p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Short Name</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {teams.map((team) => (
              <tr key={team.id}>
                <td>{team.name}</td>
                <td>{team.short_name || '—'}</td>
                <td className="actions">
                  <button className="btn btn-sm" onClick={() => handleEdit(team)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => handleDelete(team.id)}
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
