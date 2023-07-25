CREATE TABLE IF NOT EXISTS `chat_channels` (
  `id` bigint,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256),
  `name` varchar(256),
  `last_message_id` bigint,
  PRIMARY KEY (`id`),
  INDEX `idx_chat_channels_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_chat_channels_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

CREATE TABLE IF NOT EXISTS `chat_members` (
  `user_id` varchar(256),
  `channel_id` bigint,
  `last_read_message_id` bigint,
  CONSTRAINT `fk_chat_members_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_chat_members_channel` FOREIGN KEY (`channel_id`) REFERENCES `chat_channels`(`id`)
);
