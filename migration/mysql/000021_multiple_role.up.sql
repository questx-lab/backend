CREATE TABLE IF NOT EXISTS `follower_roles`(
    `created_at` datetime NULL,
    `user_id` varchar(256),
    `community_id` varchar(256),
    `role_id` varchar(256),
    PRIMARY KEY (`user_id`, `community_id`, `role_id`),
    CONSTRAINT `fk_follower_roles_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_follower_roles_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
    CONSTRAINT `fk_follower_roles_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`)
);

INSERT INTO `follower_roles`
  (`user_id`, `community_id`, `role_id`, `created_at`)
  SELECT
    `user_id`, `community_id`, `role_id`, NOW()
  FROM
    `followers`;

ALTER TABLE `followers` DROP FOREIGN KEY `fk_followers_role`;

ALTER TABLE `followers` DROP COLUMN IF EXISTS `role_id`;
