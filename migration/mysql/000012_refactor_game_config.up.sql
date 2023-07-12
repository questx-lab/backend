ALTER TABLE `game_users` DROP FOREIGN KEY `fk_game_users_game_player`;
ALTER TABLE `game_users` DROP `game_player_id`;
ALTER TABLE `game_users` ADD `character_name` varchar(256);

ALTER TABLE `game_maps` DROP `init_x`;
ALTER TABLE `game_maps` DROP `init_y`;
ALTER TABLE `game_maps` DROP `collision_layers`;
ALTER TABLE `game_maps` DROP `name`;

DROP TABLE IF EXISTS `game_map_tilesets`;
DROP TABLE IF EXISTS `game_map_players`;
