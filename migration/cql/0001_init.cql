CREATE TYPE IF NOT EXISTS attachment (
  url text
);

CREATE TYPE IF NOT EXISTS emoji (
  name text
);

CREATE TABLE IF NOT EXISTS chat_messages (
  id bigint,
  bucket bigint,
  channel_id bigint,
  author_id text,
  reply_to bigint,
  content text,
  attachments LIST<FROZEN<attachment>>,
  PRIMARY KEY ((channel_id, bucket), id)
)
WITH CLUSTERING ORDER BY (id DESC);

CREATE TABLE IF NOT EXISTS chat_reactions (
  message_id bigint,
  emoji FROZEN<emoji>,
  user_ids set<text>,
  PRIMARY KEY (message_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_chat_reactions_user_ids ON chat_reactions(user_ids);

CREATE TABLE IF NOT EXISTS channel_buckets (
  channel_id bigint,
  bucket bigint,
  quantity counter,
  PRIMARY KEY (channel_id, bucket)
)
WITH CLUSTERING ORDER BY (bucket DESC);
