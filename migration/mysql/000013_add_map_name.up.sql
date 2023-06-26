ALTER TABLE `game_maps` ADD `name` varchar(256);
CREATE UNIQUE INDEX `idx_game_maps_name` ON `game_maps` (`name`);  
