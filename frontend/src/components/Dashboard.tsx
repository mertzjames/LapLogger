import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import type { Team, Swimmer, Meet } from '../types';
import { listTeams, listSwimmers, listMeets } from '../services/api';

export default function Dashboard() {
  const { leagueId } = useParams<{ leagueId: string }>();
  const [teams, setTeams] = useState<Team[]>([]);
  const [swimmers, setSwimmers] = useState<Swimmer[]>([]);
  const [meets, setMeets] = useState<Meet[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!leagueId) return;
    Promise.all([
      listTeams(leagueId),
      listSwimmers(leagueId),
      listMeets(leagueId),
    ])
      .then(([t, s, m]) => {
        setTeams(t);
        setSwimmers(s);
        setMeets(m);
      })
      .finally(() => setLoading(false));
  }, [leagueId]);

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>
      <div className="stats-grid">
        <Link to={`/leagues/${leagueId}/teams`} className="stat-card">
          <span className="stat-number">{teams.length}</span>
          <span className="stat-label">Teams</span>
        </Link>
        <Link to={`/leagues/${leagueId}/swimmers`} className="stat-card">
          <span className="stat-number">{swimmers.length}</span>
          <span className="stat-label">Swimmers</span>
        </Link>
        <Link to={`/leagues/${leagueId}/meets`} className="stat-card">
          <span className="stat-number">{meets.length}</span>
          <span className="stat-label">Meets</span>
        </Link>
      </div>

      {meets.length > 0 && (
        <section>
          <h2>Recent Meets</h2>
          <ul className="item-list">
            {meets.slice(0, 5).map((meet) => (
              <li key={meet.id}>
                <Link to={`/leagues/${leagueId}/meets/${meet.id}/events`}>
                  <strong>{meet.name}</strong>
                  <span className="meta">
                    {meet.meet_date} &middot; {meet.location}
                    {meet.is_public && ' · Public'}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
