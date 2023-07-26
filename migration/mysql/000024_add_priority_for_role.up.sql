ALTER TABLE `roles`
  ADD COLUMN IF NOT EXISTS `priority` smallint;

ALTER TABLE `roles`
  ADD COLUMN IF NOT EXISTS `color` smallint;

ALTER TABLE `roles`
  ADD CONSTRAINT `unique_roles_community_id_priority` UNIQUE (`community_id`, `priority`);

