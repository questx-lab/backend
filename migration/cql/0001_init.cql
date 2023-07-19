CREATE TABLE IF NOT EXISTS messages (
  id bigint,
  channel_id bigint,
  user_id text,
  reply_to bigint,
  bucket bigint,
  message text,
  created_at timestamp,
  PRIMARY KEY ((channel_id,bucket), id)
)
WITH CLUSTERING ORDER BY (id DESC);

CREATE TABLE IF NOT EXISTS message_reaction_statistics (
  reaction_id text,
  message_id text,
  quantity counter,
  PRIMARY KEY (message_id, reaction_id)
)
WITH CLUSTERING ORDER BY (reaction_id DESC);

CREATE TABLE IF NOT EXISTS message_buckets (
  channel_id bigint,
  bucket bigint,
  quantity bigint,
  PRIMARY KEY (channel_id, bigint)
)
WITH CLUSTERING ORDER BY (bucket DESC);

