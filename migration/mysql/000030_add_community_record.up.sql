CREATE TABLE IF NOT EXISTS `community_stats` (
  `community_id` varchar(256),
  `date` datetime,
  `follower_count` bigint,
  PRIMARY KEY (`community_id`, `date`),
  CONSTRAINT `fk_community_stats_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);
