CREATE TABLE IF NOT EXISTS messages (
  id text,
  channel_id text,
  user_id text,
  reply_to text,
  bucket bigint,
  message text,
  created_at timestamp,
  PRIMARY KEY ((channel_id,bucket), id)
);

CREATE TABLE IF NOT EXISTS message_reaction_statistics (
  reaction_id text,
  message_id text,
  quantity counter,
  PRIMARY KEY (message_id, reaction_id)
);