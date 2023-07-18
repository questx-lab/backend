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
  content text,
  attachments LIST<FROZEN<attachment>>,
  PRIMARY KEY ((channel_id, bucket), id)
)
WITH CLUSTERING ORDER BY (id DESC);

CREATE TABLE IF NOT EXISTS chat_message_reactions (
  user_id text,
  message_id bigint,
  emoji FROZEN<emoji>,
  PRIMARY KEY (message_id, message_id)
)
WITH CLUSTERING ORDER BY (message_id DESC);

CREATE TABLE IF NOT EXISTS chat_message_reaction_statistics (
  message_id bigint,
  emoji FROZEN<emoji>,
  count int,
  PRIMARY KEY (message_id, message_id)
);
WITH CLUSTERING ORDER BY (message_id DESC);
