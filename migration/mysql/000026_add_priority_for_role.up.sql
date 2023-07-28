ALTER TABLE `roles`
  ADD COLUMN IF NOT EXISTS `priority` smallint;

ALTER TABLE `roles`
  ADD COLUMN IF NOT EXISTS `color` varchar(256);

