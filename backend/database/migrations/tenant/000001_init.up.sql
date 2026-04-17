CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    short_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE swimmers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    gender CHAR(1) NOT NULL CHECK (gender IN ('M', 'F')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE meets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    meet_date DATE NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meet_id UUID NOT NULL REFERENCES meets(id) ON DELETE CASCADE,
    stroke TEXT NOT NULL,
    distance INTEGER NOT NULL,
    unit TEXT NOT NULL DEFAULT 'yards' CHECK (unit IN ('yards', 'meters')),
    gender CHAR(1) NOT NULL DEFAULT 'X' CHECK (gender IN ('M', 'F', 'X')),
    age_group TEXT NOT NULL DEFAULT 'Open',
    is_custom BOOLEAN NOT NULL DEFAULT false,
    custom_name TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE times (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    swimmer_id UUID NOT NULL REFERENCES swimmers(id) ON DELETE CASCADE,
    time_hundredths INTEGER NOT NULL,
    is_exhibition BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, swimmer_id)
);
