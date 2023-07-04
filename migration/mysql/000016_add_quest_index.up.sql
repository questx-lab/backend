ALTER TABLE `quests` DROP IF EXISTS `index`;
ALTER TABLE `quests` ADD COLUMN `position` INT AUTO_INCREMENT UNIQUE FIRST;
ALTER TABLE `quests` CHANGE `position` `position` INT NOT NULL;
ALTER TABLE `quests` DROP INDEX `position`;
