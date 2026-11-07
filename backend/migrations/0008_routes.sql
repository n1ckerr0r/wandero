CREATE TABLE route (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL CHECK (btrim(title) <> ''),
  description TEXT NOT NULL DEFAULT '',
  published BOOLEAN NOT NULL DEFAULT false,
  distance_meters INTEGER NULL CHECK (distance_meters >= 0),
  walking_minutes INTEGER NULL CHECK (walking_minutes >= 0),
  completion_count INTEGER NOT NULL DEFAULT 0 CHECK (completion_count >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE route_stop (
  route_id BIGINT NOT NULL REFERENCES route(id) ON DELETE CASCADE,
  position INTEGER NOT NULL CHECK (position > 0),
  location_id BIGINT NOT NULL REFERENCES location(id) ON DELETE RESTRICT,
  visit_minutes INTEGER NOT NULL DEFAULT 30 CHECK (visit_minutes >= 0 AND visit_minutes <= 1440),
  PRIMARY KEY (route_id, position),
  UNIQUE (route_id, location_id)
);

CREATE INDEX route_published_updated_at_idx ON route (published, updated_at DESC);
CREATE INDEX route_stop_location_id_idx ON route_stop (location_id);
