import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import type { SwimEvent } from '../types';
import {
  listEvents,
  createEvent,
  updateEvent,
  deleteEvent,
} from '../services/api';

const STROKES = ['Free', 'Back', 'Breast', 'Fly', 'IM'];
const DISTANCES = [25, 50, 100, 200, 400, 500, 1650];
const AGE_GROUPS = ['8&U', '9-10', '11-12', '13-14', '15-18', 'Open'];

const emptyForm = {
  stroke: 'Free',
  distance: 50,
  unit: 'yards',
  gender: 'X',
  age_group: 'Open',
  is_custom: false,
  custom_name: '',
  sort_order: 0,
};

export default function Events() {
  const { leagueId, meetId } = useParams<{
    leagueId: string;
    meetId: string;
  }>();
  const [events, setEvents] = useState<SwimEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);

  const load = () => {
    if (!leagueId || !meetId) return;
    listEvents(leagueId, meetId)
      .then(setEvents)
      .catch(() => setError('Failed to load events'))
      .finally(() => setLoading(false));
  };

  useEffect(load, [leagueId, meetId]);

  const resetForm = () => {
    setForm(emptyForm);
    setEditingId(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leagueId || !meetId) return;
    setError('');
    try {
      if (editingId) {
        await updateEvent(leagueId, meetId, editingId, form);
      } else {
        await createEvent(leagueId, meetId, form);
      }
      resetForm();
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save event');
    }
  };

  const handleEdit = (ev: SwimEvent) => {
    setEditingId(ev.id);
    setForm({
      stroke: ev.stroke,
      distance: ev.distance,
      unit: ev.unit,
      gender: ev.gender,
      age_group: ev.age_group,
      is_custom: ev.is_custom,
      custom_name: ev.custom_name || '',
      sort_order: ev.sort_order,
    });
  };

  const handleDelete = async (id: string) => {
    if (!leagueId || !meetId) return;
    try {
      await deleteEvent(leagueId, meetId, id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  const eventLabel = (ev: SwimEvent) =>
    ev.is_custom && ev.custom_name
      ? ev.custom_name
      : `${ev.distance} ${ev.unit} ${ev.stroke}`;

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="events-page">
      <h1>Events</h1>
      {error && <div className="error-msg">{error}</div>}

      <form onSubmit={handleSubmit} className="crud-form">
        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={form.is_custom}
            onChange={(e) =>
              setForm({ ...form, is_custom: e.target.checked })
            }
          />
          Custom Event
        </label>

        {form.is_custom ? (
          <input
            type="text"
            placeholder="Custom event name"
            value={form.custom_name}
            onChange={(e) => setForm({ ...form, custom_name: e.target.value })}
            required
          />
        ) : (
          <select
            value={form.stroke}
            onChange={(e) => setForm({ ...form, stroke: e.target.value })}
          >
            {STROKES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        )}

        <select
          value={form.distance}
          onChange={(e) =>
            setForm({ ...form, distance: parseInt(e.target.value, 10) })
          }
        >
          {DISTANCES.map((d) => (
            <option key={d} value={d}>{d}</option>
          ))}
        </select>

        <select
          value={form.unit}
          onChange={(e) => setForm({ ...form, unit: e.target.value })}
        >
          <option value="yards">Yards</option>
          <option value="meters">Meters</option>
        </select>

        <select
          value={form.gender}
          onChange={(e) => setForm({ ...form, gender: e.target.value })}
        >
          <option value="X">Mixed</option>
          <option value="M">Male</option>
          <option value="F">Female</option>
        </select>

        <select
          value={form.age_group}
          onChange={(e) => setForm({ ...form, age_group: e.target.value })}
        >
          {AGE_GROUPS.map((ag) => (
            <option key={ag} value={ag}>{ag}</option>
          ))}
        </select>

        <input
          type="number"
          placeholder="Sort order"
          value={form.sort_order}
          onChange={(e) =>
            setForm({ ...form, sort_order: parseInt(e.target.value, 10) || 0 })
          }
        />

        <div className="form-actions">
          <button type="submit" className="btn btn-primary">
            {editingId ? 'Update' : 'Add Event'}
          </button>
          {editingId && (
            <button type="button" className="btn" onClick={resetForm}>
              Cancel
            </button>
          )}
        </div>
      </form>

      {events.length === 0 ? (
        <p className="empty-state">No events yet.</p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Event</th>
              <th>Gender</th>
              <th>Age Group</th>
              <th>Times</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {events.map((ev) => (
              <tr key={ev.id}>
                <td>{ev.sort_order}</td>
                <td>{eventLabel(ev)}</td>
                <td>{ev.gender}</td>
                <td>{ev.age_group}</td>
                <td>
                  <Link
                    to={`/leagues/${leagueId}/meets/${meetId}/events/${ev.id}/times`}
                    className="btn btn-sm"
                  >
                    Times
                  </Link>
                </td>
                <td className="actions">
                  <button className="btn btn-sm" onClick={() => handleEdit(ev)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => handleDelete(ev.id)}
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
