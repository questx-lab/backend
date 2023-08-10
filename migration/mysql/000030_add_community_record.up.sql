CREATE TABLE IF NOT EXISTS `community_records` (
  `community_id` varchar(256),
  `date` datetime,
  `followers` bigint,
  PRIMARY KEY (`community_id`, `date`),
  CONSTRAINT `fk_community_records_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);
