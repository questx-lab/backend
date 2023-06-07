DROP TABLE IF EXISTS `badges` CASCADE;
CREATE TABLE `badges` (
    `id` varchar(256),
    `created_at` datetime NULL,
    `updated_at` datetime NULL,
    `deleted_at` datetime NULL,
    `name` varchar(256),
    `level` bigint,
    `description` varchar(256),
    `value` bigint,
    `icon_url` varchar(256),
    PRIMARY KEY (`id`),
    INDEX `idx_badges_deleted_at` (`deleted_at`),
    UNIQUE INDEX `idx_badges_name_level` (`name`,`level`)
);

CREATE TABLE `badge_details` (
    `user_id` varchar(256),
    `community_id` varchar(256),
    `badge_id` varchar(256),
    `was_notified` boolean,
    `created_at` datetime NULL,
    PRIMARY KEY (
        `user_id`, `community_id`, `badge_id`
    ), 
    CONSTRAINT `fk_badge_details_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`), 
    CONSTRAINT `fk_badge_details_badge` FOREIGN KEY (`badge_id`) REFERENCES `badges`(`id`), 
    CONSTRAINT `fk_badge_details_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

