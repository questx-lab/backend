ALTER TABLE `game_luckyboxes` DROP COLUMN `is_random`;
ALTER TABLE `game_luckyboxes` ADD `collected_at` datetime NULL;
