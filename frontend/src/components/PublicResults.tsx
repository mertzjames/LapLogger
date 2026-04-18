import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { PublicMeetResult } from '../types';
import { formatTime } from '../types';
import { getPublicResults, ApiError } from '../services/api';

export default function PublicResults() {
  const { leagueSlug, meetId } = useParams<{
    leagueSlug: string;
    meetId: string;
  }>();
  const [result, setResult] = useState<PublicMeetResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!leagueSlug || !meetId) return;
    getPublicResults(leagueSlug, meetId)
      .then(setResult)
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          setError('Meet not found or not public.');
        } else {
          setError('Failed to load results.');
        }
      })
      .finally(() => setLoading(false));
  }, [leagueSlug, meetId]);

  if (loading) return <div className="loading">Loading results...</div>;
  if (error) return <div className="public-results"><div className="error-msg">{error}</div></div>;
  if (!result) return null;

  return (
    <div className="public-results">
      <header className="results-header">
        <h1>{result.meet_name}</h1>
        <p className="meta">
          {result.league_name} &middot; {result.location} &middot;{' '}
          {result.meet_date}
        </p>
      </header>

      {result.events.length === 0 ? (
        <p className="empty-state">No events recorded for this meet.</p>
      ) : (
        result.events.map((event) => {
          const label =
            event.is_custom && event.custom_name
              ? event.custom_name
              : `${event.distance} ${event.unit} ${event.stroke}`;

          return (
            <section key={event.id} className="result-event">
              <h2>
                {label}
                <span className="event-meta">
                  {event.gender !== 'X' && ` · ${event.gender === 'M' ? 'Boys' : 'Girls'}`}
                  {event.age_group !== 'Open' && ` · ${event.age_group}`}
                </span>
              </h2>

              {event.times.length === 0 ? (
                <p className="empty-state">No times recorded.</p>
              ) : (
                <table className="data-table results-table">
                  <thead>
                    <tr>
                      <th>Place</th>
                      <th>Swimmer</th>
                      <th>Team</th>
                      <th>Time</th>
                    </tr>
                  </thead>
                  <tbody>
                    {event.times.map((t, i) => (
                      <tr
                        key={`${event.id}-${i}`}
                        className={t.is_exhibition ? 'exhibition' : ''}
                      >
                        <td>{t.is_exhibition ? 'EX' : i + 1}</td>
                        <td>{t.swimmer_name}</td>
                        <td>{t.team_name || '—'}</td>
                        <td className="time-value">
                          {formatTime(t.time_hundredths)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </section>
          );
        })
      )}
    </div>
  );
}
