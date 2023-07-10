ALTER TABLE `game_users` DROP COLUMN IF EXISTS `is_active`;
ALTER TABLE `game_users` ADD COLUMN IF NOT EXISTS `connected_by` VARCHAR(256) NULL;
UPDATE `game_users` SET `character_id` = NULL;
