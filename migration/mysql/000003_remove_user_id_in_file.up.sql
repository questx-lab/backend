ALTER TABLE `files` DROP CONSTRAINT `fk_files_user`;
ALTER TABLE `files` DROP COLUMN `user_id`;
ALTER TABLE `files` ADD CONSTRAINT `fk_files_created_by_user` FOREIGN KEY(`created_by`) REFERENCES `users`(`id`);
