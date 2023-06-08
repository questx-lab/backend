ALTER TABLE `users` CHANGE `referral_code` `invite_code` varchar(256) UNIQUE;
ALTER TABLE `followers` DROP COLUMN `invite_code`;
