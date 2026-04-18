import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import type { League } from '../types';
import { listLeagues, createLeague } from '../services/api';

export default function LeagueSelect() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [newName, setNewName] = useState('');
  const [creating, setCreating] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    listLeagues()
      .then(setLeagues)
      .catch(() => setError('Failed to load leagues'))
      .finally(() => setLoading(false));
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    setCreating(true);
    setError('');
    try {
      const league = await createLeague(newName.trim());
      navigate(`/leagues/${league.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create league');
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <div className="loading">Loading leagues...</div>;

  return (
    <div className="league-select">
      <h1>Your Leagues</h1>
      {error && <div className="error-msg">{error}</div>}

      {leagues.length > 0 ? (
        <ul className="league-list">
          {leagues.map((league) => (
            <li key={league.id}>
              <button
                className="league-item"
                onClick={() => navigate(`/leagues/${league.id}`)}
              >
                <span className="league-name">{league.name}</span>
                <span className="league-slug">{league.slug}</span>
              </button>
            </li>
          ))}
        </ul>
      ) : (
        <p className="empty-state">No leagues yet. Create one below.</p>
      )}

      <form onSubmit={handleCreate} className="create-form">
        <h2>Create a League</h2>
        <div className="form-row">
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="League name"
            required
          />
          <button type="submit" className="btn btn-primary" disabled={creating}>
            {creating ? 'Creating...' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  );
}
