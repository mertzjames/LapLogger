import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { TimeEntry as TimeEntryType, Swimmer } from '../types';
import { formatTime, parseTime } from '../types';
import {
  listTimes,
  listSwimmers,
  createTime,
  updateTime,
  deleteTime,
} from '../services/api';

export default function TimeEntry() {
  const { leagueId, meetId, eventId } = useParams<{
    leagueId: string;
    meetId: string;
    eventId: string;
  }>();
  const [times, setTimes] = useState<TimeEntryType[]>([]);
  const [swimmers, setSwimmers] = useState<Swimmer[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formSwimmerId, setFormSwimmerId] = useState('');
  const [formTimeStr, setFormTimeStr] = useState('');
  const [formExhibition, setFormExhibition] = useState(false);

  const load = () => {
    if (!leagueId || !meetId || !eventId) return;
    Promise.all([
      listTimes(leagueId, meetId, eventId),
      listSwimmers(leagueId),
    ])
      .then(([t, s]) => {
        setTimes(t);
        setSwimmers(s);
      })
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false));
  };

  useEffect(load, [leagueId, meetId, eventId]);

  const resetForm = () => {
    setFormSwimmerId('');
    setFormTimeStr('');
    setFormExhibition(false);
    setEditingId(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leagueId || !meetId || !eventId) return;
    const hundredths = parseTime(formTimeStr);
    if (hundredths === null) {
      setError('Invalid time format. Use MM:SS.ms or SS.ms (e.g., 1:05.23 or 25.34)');
      return;
    }
    setError('');
    try {
      if (editingId) {
        await updateTime(leagueId, meetId, eventId, editingId, {
          swimmer_id: formSwimmerId,
          time_hundredths: hundredths,
          is_exhibition: formExhibition,
        });
      } else {
        await createTime(leagueId, meetId, eventId, {
          swimmer_id: formSwimmerId,
          time_hundredths: hundredths,
          is_exhibition: formExhibition,
        });
      }
      resetForm();
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save time');
    }
  };

  const handleEdit = (t: TimeEntryType) => {
    setEditingId(t.id);
    setFormSwimmerId(t.swimmer_id);
    setFormTimeStr(formatTime(t.time_hundredths));
    setFormExhibition(t.is_exhibition);
  };

  const handleDelete = async (id: string) => {
    if (!leagueId || !meetId || !eventId) return;
    try {
      await deleteTime(leagueId, meetId, eventId, id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  const swimmerName = (id: string) => {
    const s = swimmers.find((sw) => sw.id === id);
    return s ? `${s.first_name} ${s.last_name}` : id;
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="time-entry-page">
      <h1>Time Entry</h1>
      {error && <div className="error-msg">{error}</div>}

      <form onSubmit={handleSubmit} className="crud-form">
        <select
          value={formSwimmerId}
          onChange={(e) => setFormSwimmerId(e.target.value)}
          required
        >
          <option value="">Select Swimmer</option>
          {swimmers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.first_name} {s.last_name}
            </option>
          ))}
        </select>
        <input
          type="text"
          placeholder="Time (e.g., 1:05.23)"
          value={formTimeStr}
          onChange={(e) => setFormTimeStr(e.target.value)}
          required
        />
        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={formExhibition}
            onChange={(e) => setFormExhibition(e.target.checked)}
          />
          Exhibition
        </label>
        <div className="form-actions">
          <button type="submit" className="btn btn-primary">
            {editingId ? 'Update' : 'Add Time'}
          </button>
          {editingId && (
            <button type="button" className="btn" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      {times.length === 0 ? (
        <p className="empty-state">No times recorded yet.</p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>Place</th>
              <th>Swimmer</th>
              <th>Time</th>
              <th>Exhibition</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {times.map((t, i) => (
              <tr key={t.id} className={t.is_exhibition ? 'exhibition' : ''}>
                <td>{t.is_exhibition ? '—' : i + 1}</td>
                <td>{swimmerName(t.swimmer_id)}</td>
                <td className="time-value">{formatTime(t.time_hundredths)}</td>
                <td>{t.is_exhibition ? 'Yes' : 'No'}</td>
                <td className="actions">
                  <button className="btn btn-sm" onClick={() => handleEdit(t)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => handleDelete(t.id)}
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
