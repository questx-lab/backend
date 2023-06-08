ALTER TABLE `users` CHANGE `referral_code` `invite_code` varchar(256) UNIQUE;
ALTER TABLE `followers` DROP COLUMN `invite_code`;
ALTER TABLE `communities` CHANGE `referral_status` `invited_status` varchar(256);
ALTER TABLE `communities` CHANGE `referred_by` `invited_by` varchar(256);
