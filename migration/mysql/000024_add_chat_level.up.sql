ALTER TABLE `followers` ADD COLUMN IF NOT EXISTS `total_chat_xp` bigint NOT NULL;
ALTER TABLE `followers` ADD COLUMN IF NOT EXISTS `current_chat_xp` bigint NOT NULL;
ALTER TABLE `followers` ADD COLUMN IF NOT EXISTS `chat_level` bigint NOT NULL;
