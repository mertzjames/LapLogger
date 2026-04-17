-- Add indexes on foreign key columns for query performance
CREATE INDEX IF NOT EXISTS idx_swimmers_team_id ON swimmers(team_id);
CREATE INDEX IF NOT EXISTS idx_events_meet_id ON events(meet_id);
CREATE INDEX IF NOT EXISTS idx_times_event_id ON times(event_id);
CREATE INDEX IF NOT EXISTS idx_times_swimmer_id ON times(swimmer_id);
