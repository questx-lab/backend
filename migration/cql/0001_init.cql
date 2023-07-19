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
  user_id text,
  message_id bigint,
  emoji FROZEN<emoji>,
  PRIMARY KEY ((message_id, emoji), user_id)
);

CREATE TABLE IF NOT EXISTS chat_reaction_statistics (
  message_id bigint,
  emoji FROZEN<emoji>,
  count counter,
  PRIMARY KEY (message_id, emoji)
);

CREATE TABLE IF NOT EXISTS message_buckets (
  channel_id bigint,
  bucket bigint,
  quantity bigint,
  PRIMARY KEY (channel_id, bucket)
)
WITH CLUSTERING ORDER BY (bucket DESC);
