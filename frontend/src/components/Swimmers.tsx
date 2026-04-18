import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { Swimmer, Team } from '../types';
import {
  listSwimmers,
  listTeams,
  createSwimmer,
  updateSwimmer,
  deleteSwimmer,
} from '../services/api';

const emptyForm = {
  team_id: '',
  first_name: '',
  last_name: '',
  date_of_birth: '',
  gender: 'M',
};

export default function Swimmers() {
  const { leagueId } = useParams<{ leagueId: string }>();
  const [swimmers, setSwimmers] = useState<Swimmer[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [filterTeam, setFilterTeam] = useState('');

  const load = () => {
    if (!leagueId) return;
    Promise.all([
      listSwimmers(leagueId, filterTeam || undefined),
      listTeams(leagueId),
    ])
      .then(([s, t]) => {
        setSwimmers(s);
        setTeams(t);
      })
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false));
  };

  useEffect(load, [leagueId, filterTeam]);

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
        await updateSwimmer(leagueId, editingId, form);
      } else {
        await createSwimmer(leagueId, form);
      }
      resetForm();
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save swimmer');
    }
  };

  const handleEdit = (s: Swimmer) => {
    setEditingId(s.id);
    setForm({
      team_id: s.team_id,
      first_name: s.first_name,
      last_name: s.last_name,
      date_of_birth: s.date_of_birth,
      gender: s.gender,
    });
  };

  const handleDelete = async (id: string) => {
    if (!leagueId) return;
    setError('');
    try {
      await deleteSwimmer(leagueId, id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  const teamName = (teamId: string) =>
    teams.find((t) => t.id === teamId)?.name || '—';

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="swimmers-page">
      <h1>Swimmers</h1>
      {error && <div className="error-msg">{error}</div>}

      <div className="filter-bar">
        <select
          value={filterTeam}
          onChange={(e) => setFilterTeam(e.target.value)}
        >
          <option value="">All Teams</option>
          {teams.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
      </div>

      <form onSubmit={handleSubmit} className="crud-form">
        <select
          value={form.team_id}
          onChange={(e) => setForm({ ...form, team_id: e.target.value })}
          required
        >
          <option value="">Select Team</option>
          {teams.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
        <input
          type="text"
          placeholder="First name"
          value={form.first_name}
          onChange={(e) => setForm({ ...form, first_name: e.target.value })}
          required
        />
        <input
          type="text"
          placeholder="Last name"
          value={form.last_name}
          onChange={(e) => setForm({ ...form, last_name: e.target.value })}
          required
        />
        <input
          type="date"
          value={form.date_of_birth}
          onChange={(e) => setForm({ ...form, date_of_birth: e.target.value })}
          required
        />
        <select
          value={form.gender}
          onChange={(e) => setForm({ ...form, gender: e.target.value })}
        >
          <option value="M">Male</option>
          <option value="F">Female</option>
        </select>
        <div className="form-actions">
          <button type="submit" className="btn btn-primary">
            {editingId ? 'Update' : 'Add Swimmer'}
          </button>
          {editingId && (
            <button type="button" className="btn" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      {swimmers.length === 0 ? (
        <p className="empty-state">No swimmers yet.</p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Team</th>
              <th>DOB</th>
              <th>Gender</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {swimmers.map((s) => (
              <tr key={s.id}>
                <td>
                  {s.first_name} {s.last_name}
                </td>
                <td>{teamName(s.team_id)}</td>
                <td>{s.date_of_birth}</td>
                <td>{s.gender}</td>
                <td className="actions">
                  <button className="btn btn-sm" onClick={() => handleEdit(s)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => handleDelete(s.id)}
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
