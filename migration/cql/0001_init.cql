CREATE TABLE IF NOT EXISTS messages (
  id text,
  channel_id text,
  content text,
  reaction_counts map<text, int>,
  created_at timestamp,
  PRIMARY KEY (channel_id, created_at)
)
WITH CLUSTERING ORDER BY (created_at DESC);

CREATE TABLE IF NOT EXISTS user_reactions (
  user_id text,
  reaction_id text,
  message_id text,
  created_at timestamp,
  PRIMARY KEY (message_id, created_at)
)
WITH CLUSTERING ORDER BY (created_at DESC);