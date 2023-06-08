CREATE TABLE IF NOT EXISTS `users`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `wallet_address` varchar(256) UNIQUE,
    `name` varchar(256) UNIQUE,
    `role` varchar(256),
    `profile_picture` varchar(256),
    `referral_code` varchar(256),
    `is_new_user` boolean,
    PRIMARY KEY (`id`),
    INDEX `idx_users_deleted_at`(`deleted_at`)
);

CREATE TABLE IF NOT EXISTS `communities` (
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `created_by` varchar(256),
    `referred_by` varchar(256),
    `referral_status` varchar(256),
    `handle` varchar(256) UNIQUE,
    `display_name` varchar(256),
    `followers` bigint,
    `trending_score` bigint,
    `logo_picture` varchar(256),
    `introduction` longtext,
    `twitter` varchar(256),
    `discord` varchar(256),
    `website_url` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_communities_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_communities_created_by_user` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_communities_referred_by_user` FOREIGN KEY (`referred_by`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `categories`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `name` varchar(256) UNIQUE,
    `community_id` varchar(256),
    `created_by` varchar(256) NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_categories_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_categories_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
    CONSTRAINT `fk_categories_created_by_user` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `quests`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `community_id` varchar(256),
    `is_template` boolean,
    `type` varchar(256),
    `status` varchar(256),
    `index` bigint,
    `title` varchar(256),
    `description` longtext,
    `category_id` varchar(256),
    `recurrence` varchar(256),
    `validation_data` longblob,
    `points` bigint unsigned,
    `rewards` longblob,
    `condition_op` varchar(256),
    `conditions` longblob,
    `is_highlight` boolean,
    PRIMARY KEY (`id`),
    INDEX `idx_quests_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_quests_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
    CONSTRAINT `fk_quests_category` FOREIGN KEY (`category_id`) REFERENCES `categories`(`id`)
);

CREATE TABLE IF NOT EXISTS `collaborators`(
    `user_id` varchar(256),
    `community_id` varchar(256),
    `role` varchar(256),
    `created_by` varchar(256),
    PRIMARY KEY (`user_id`, `community_id`),
    CONSTRAINT `fk_collaborators_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_collaborators_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
    CONSTRAINT `fk_collaborators_created_by_user` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `claimed_quests`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `quest_id` varchar(256),
    `user_id` varchar(256),
    `submission_data` varchar(256),
    `status` varchar(256),
    `reviewer_id` varchar(256),
    `reviewed_at` datetime NULL,
    `comment` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_claimed_quests_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_claimed_quests_quest` FOREIGN KEY (`quest_id`) REFERENCES `quests`(`id`),
    CONSTRAINT `fk_claimed_quests_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `followers`(
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `user_id` varchar(256),
    `community_id` varchar(256),
    `points` bigint unsigned,
    `quests` bigint unsigned,
    `streaks` bigint unsigned,
    `invite_code` varchar(256) UNIQUE,
    `invite_count` bigint unsigned,
    `invited_by` varchar(256),
    PRIMARY KEY (`user_id`, `community_id`),
    INDEX `idx_followers_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_followers_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_followers_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
    CONSTRAINT `fk_followers_invited_by_user` FOREIGN KEY (`invited_by`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `api_keys`(
    `community_id` varchar(256),
    `key` varchar(256),
    PRIMARY KEY (`community_id`),
    INDEX `idx_api_keys_key`(`key`),
    CONSTRAINT `fk_api_keys_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

CREATE TABLE IF NOT EXISTS `refresh_tokens`(
    `user_id` varchar(256),
    `family` varchar(256) UNIQUE,
    `counter` bigint unsigned,
    `expiration` datetime NULL,
    CONSTRAINT `fk_refresh_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `files`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `mime` varchar(256),
    `name` varchar(256),
    `created_by` varchar(256) NOT NULL,
    `user_id` varchar(256),
    `url` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_files_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_files_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `badges`(
    `user_id` varchar(256),
    `community_id` varchar(256),
    `name` varchar(256),
    `level` bigint,
    `was_notified` boolean,
    PRIMARY KEY (`user_id`, `community_id`, `name`),
    CONSTRAINT `fk_badges_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_badges_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

CREATE TABLE IF NOT EXISTS `game_maps`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `name` varchar(256) UNIQUE,
    `init_x` bigint,
    `init_y` bigint,
    `map` longblob,
    `player` longblob,
    `map_path` varchar(256),
    `tile_set_path` varchar(256),
    `player_img_path` varchar(256),
    `player_json_path` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_game_maps_deleted_at`(`deleted_at`)
);

CREATE TABLE IF NOT EXISTS `game_rooms`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `name` varchar(256),
    `map_id` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_game_rooms_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_game_rooms_game_map` FOREIGN KEY (`map_id`) REFERENCES `game_maps`(`id`)
);

CREATE TABLE IF NOT EXISTS `game_users`(
    `room_id` varchar(256),
    `user_id` varchar(256),
    `direction` varchar(256),
    `position_x` bigint,
    `position_y` bigint,
    `is_active` boolean,
    PRIMARY KEY (`room_id`, `user_id`),
    CONSTRAINT `fk_game_users_game_room` FOREIGN KEY (`room_id`) REFERENCES `game_rooms`(`id`),
    CONSTRAINT `fk_game_users_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

CREATE TABLE IF NOT EXISTS `pay_rewards`(
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `user_id` varchar(256),
    `claimed_quest_id` varchar(256),
    `note` varchar(256),
    `status` varchar(256),
    `address` varchar(256),
    `token` varchar(256),
    `amount` double,
    `tx_hash` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_pay_rewards_deleted_at`(`deleted_at`),
    CONSTRAINT `fk_pay_rewards_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_pay_rewards_claimed_quest` FOREIGN KEY (`claimed_quest_id`) REFERENCES `claimed_quests`(`id`)
);

