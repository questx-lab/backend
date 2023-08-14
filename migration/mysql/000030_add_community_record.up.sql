CREATE TABLE IF NOT EXISTS `community_stats` (
  `community_id` varchar(256) NULL,
  `date` datetime NOT NULL,
  `follower_count` bigint NOT NULL,
  CONSTRAINT `fk_community_stats_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);
