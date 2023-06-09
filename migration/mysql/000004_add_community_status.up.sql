ALTER TABLE `communities` ADD `status` varchar(256);
UPDATE `communities` SET `status`='active';
