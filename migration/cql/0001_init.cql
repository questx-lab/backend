CREATE TABLE IF NOT EXISTS user (
  id text,
  name text,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS channel (
  id text,
  community_id text,
  created_by text,
  created_at time,
  updated_at time,
  deleted_at time,
  name text,
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS channel_community_id_idx ON channel (community_id);
CREATE INDEX IF NOT EXISTS channel_created_by_idx ON channel (created_by);

CREATE TABLE IF NOT EXISTS message (
  id text,
  author text,
  created_at text,
  body text,
  channel_id text,
  PRIMARY KEY (channel_id, created_at)
);