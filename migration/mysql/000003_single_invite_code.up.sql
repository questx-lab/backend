ALTER TABLE `users` CHANGE `referral_code` `invite_code` varchar(256) UNIQUE;
ALTER TABLE `followers` DROP COLUMN `invite_code`;
ALTER TABLE `communities` CHANGE `referral_status` `invited_status` varchar(256);
ALTER TABLE `communities` CHANGE `referred_by` `invited_by` varchar(256);
ALTER TABLE `communities` DROP CONSTRAINT `fk_communities_referred_by_user`;
ALTER TABLE `communities` ADD CONSTRAINT `fk_communities_invited_by_user` 
    FOREIGN KEY(`invited_by`) REFERENCES `users`(`id`);
