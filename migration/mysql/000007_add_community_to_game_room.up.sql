DROP TABLE IF EXISTS `game_users` CASCADE;
DROP TABLE IF EXISTS `game_rooms` CASCADE;
DROP TABLE IF EXISTS `game_maps` CASCADE;

CREATE TABLE IF NOT EXISTS `game_maps` (
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `name` varchar(256) UNIQUE,
    `init_x` bigint,
    `init_y` bigint,
    `config_url` varchar(256),
    `collision_layers` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_game_maps_deleted_at` (`deleted_at`)
);

CREATE TABLE IF NOT EXISTS `game_map_players` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `game_map_id` varchar(256),
  `name` varchar(256),
  `config_url` varchar(256),
  `image_url` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_game_map_players_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_map_id_name` (`game_map_id`, `name`),
  CONSTRAINT `fk_game_map_players_game_map` FOREIGN KEY (`game_map_id`) REFERENCES `game_maps`(`id`)
);

CREATE TABLE IF NOT EXISTS `game_map_tilesets` (
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `game_map_id` varchar(256),
    `tileset_url` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_game_map_tilesets_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_game_map_tilesets_game_map` FOREIGN KEY (`game_map_id`) REFERENCES `game_maps`(`id`)
);

CREATE TABLE `game_rooms` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256),
  `map_id` varchar(256),
  `name` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_game_rooms_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_game_rooms_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
  CONSTRAINT `fk_game_rooms_game_map` FOREIGN KEY (`map_id`) REFERENCES `game_maps`(`id`)
);

CREATE TABLE `game_users` (
  `room_id` varchar(256),
  `user_id` varchar(256),
  `game_player_id` varchar(256),
  `direction` varchar(256),
  `position_x` bigint,
  `position_y` bigint,
  `is_active` boolean,
  PRIMARY KEY (`room_id`, `user_id`),
  CONSTRAINT `fk_game_users_game_player` FOREIGN KEY (`game_player_id`) REFERENCES `game_map_players`(`id`),
  CONSTRAINT `fk_game_users_room` FOREIGN KEY (`room_id`) REFERENCES `game_rooms`(`id`),
  CONSTRAINT `fk_game_users_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);
