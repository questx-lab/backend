CREATE TABLE IF NOT EXISTS `roles`(
  `id` varchar(256),
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

INSERT INTO `roles`(id, name, permissions)
  VALUES (uuid(), "user", 0),
(uuid(), "owner",(1 << 63) - 1),
(uuid(), "editor",(1 << 1) +(1 << 2)),
(uuid(), "reviewer",(1 << 3));

UPDATE
  `followers`
SET
  role_id =(
    SELECT
      id
    FROM
      roles
    WHERE
      name = 'user');

INSERT INTO `followers`(user_id, community_id, points, streaks, quests, invite_count, created_at, updated_at, role_id)
SELECT
  user_id,
  community_id,
  0,
  0,
  0,
  0,
  NOW(),
  NOW(),
(
    SELECT
      id
    FROM
      roles
    WHERE
      name = collaborators.role) AS role_id
FROM
  `collaborators` ON DUPLICATE KEY UPDATE
    role_id =(
      SELECT
        id
      FROM
        roles
      WHERE
        name = collaborators.role), updated_at = NOW();

DROP TABLE IF EXISTS `collaborators` CASCADE;

