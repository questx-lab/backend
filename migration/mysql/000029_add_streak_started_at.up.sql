ALTER TABLE `followers` DROP IF EXISTS `streaks`;

CREATE TABLE IF NOT EXISTS `follower_streaks` (
  `user_id` varchar(256),
  `community_id` varchar(256),
  `start_time` datetime,
  `streaks` bigint,
  PRIMARY KEY (`user_id`, `community_id`, `start_time`), 
  CONSTRAINT `fk_follower_streaks_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_follower_streaks_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);
