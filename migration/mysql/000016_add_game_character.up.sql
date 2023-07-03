ALTER TABLE `game_users` CHANGE `character_name` `character_id` varchar(256) NULL;
UPDATE `game_users` SET `character_id`=null;

CREATE TABLE IF NOT EXISTS `game_characters` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(256),
  `level` bigint,
  `config_url` varchar(256),
  `image_url` varchar(256),
  `sprite_width_ratio` double,
  `sprite_height_ratio` double,
  PRIMARY KEY (`id`),
  INDEX `idx_game_characters_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_game_characters_name_level` (`name`, `level`)
);

CREATE TABLE IF NOT EXISTS `game_community_characters` (
  `community_id` varchar(256),
  `character_id` varchar(256),
  `points` bigint,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (`community_id`, `character_id`),
  CONSTRAINT `fk_game_community_characters_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
  CONSTRAINT `fk_game_community_characters_character` FOREIGN KEY (`character_id`) REFERENCES `game_characters`(`id`)
);

CREATE TABLE IF NOT EXISTS `game_user_characters` (
  `user_id` varchar(256),
  `community_id` varchar(256),
  `character_id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  PRIMARY KEY (
    `user_id`, `community_id`, `character_id`
  ),
  CONSTRAINT `fk_game_user_characters_character` FOREIGN KEY (`character_id`) REFERENCES `game_characters`(`id`),
  CONSTRAINT `fk_game_user_characters_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_game_user_characters_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);
