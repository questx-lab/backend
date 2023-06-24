ALTER TABLE `game_luckyboxes` DROP COLUMN IF EXISTS `is_random`;
ALTER TABLE `game_luckyboxes` ADD `collected_at` datetime NULL;
