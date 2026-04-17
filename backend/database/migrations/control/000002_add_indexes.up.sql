-- league_memberships already has a composite PK on (user_id, league_id),
-- which serves as the primary index. Add an explicit index for lookups
-- by user_id alone (used by GetUserLeagues).
CREATE INDEX IF NOT EXISTS idx_league_memberships_user_id ON league_memberships(user_id);
