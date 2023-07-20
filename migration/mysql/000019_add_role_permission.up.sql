CREATE TABLE IF NOT EXISTS `roles`(
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256) NULL,
  `name` varchar(256),
  `permissions` bigint unsigned,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_roles_community_id_name`(`community_id`, `name`),
  CONSTRAINT `fk_roles_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

ALTER TABLE `followers`
  ADD COLUMN IF NOT EXISTS `role_id` varchar(256);

ALTER TABLE `followers`
  ADD CONSTRAINT `fk_followers_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`);

INSERT INTO `roles`(`id`, `name`, `permissions`)
  VALUES ("user", "user", 0),
         ("owner", "owner", (1 << 63) - 1)
  ON DUPLICATE KEY UPDATE `id`=`id`;

UPDATE `followers` SET `role_id`="user";

INSERT INTO `followers`
  (`user_id`, `community_id`, `points`, `streaks`, `quests`, `invite_count`, 
   `created_at`, `updated_at`, `role_id`)
  SELECT
    `user_id`, `community_id`, 0, 0, 0, 0, NOW(), NOW(), "owner"
  FROM
    `collaborators`
  WHERE
    `role`="owner"
ON DUPLICATE KEY UPDATE `role_id` = "owner", `updated_at` = NOW();

DROP TABLE IF EXISTS `collaborators` CASCADE;

ALTER TABLE `followers` 
  DROP PRIMARY KEY, ADD PRIMARY KEY(`user_id`, `community_id`, `role_id`);
