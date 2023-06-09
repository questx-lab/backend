ALTER TABLE `communities` ADD `owner_email` varchar(256);
ALTER TABLE `communities` ADD `status` varchar(256);
UPDATE `communities` SET `status`='active';
