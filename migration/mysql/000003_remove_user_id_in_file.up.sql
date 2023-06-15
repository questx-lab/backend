ALTER TABLE `files` DROP FOREIGN KEY `fk_files_user`;
ALTER TABLE `files` DROP COLUMN IF EXISTS `user_id`;
ALTER TABLE `files` ADD CONSTRAINT `fk_files_created_by_user` FOREIGN KEY(`created_by`) REFERENCES `users`(`id`);
