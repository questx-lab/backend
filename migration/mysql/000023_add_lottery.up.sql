CREATE TABLE IF NOT EXISTS `lottery_events` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256),
  `start_time` datetime NULL,
  `end_time` datetime NULL,
  `max_tickets` bigint,
  `used_tickets` bigint,
  `point_per_ticket` bigint unsigned,
  PRIMARY KEY (`id`),
  INDEX `idx_lottery_events_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_lottery_events_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

CREATE TABLE IF NOT EXISTS `lottery_prizes` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `lottery_event_id` varchar(256),
  `points` bigint,
  `rewards` longblob,
  `available_rewards` bigint,
  `won_rewards` bigint,
  PRIMARY KEY (`id`),
  INDEX `idx_lottery_prizes_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_lottery_prizes_lottery_event` FOREIGN KEY (`lottery_event_id`) REFERENCES `lottery_events`(`id`)
);

CREATE TABLE IF NOT EXISTS `lottery_winners` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `lottery_prize_id` varchar(256),
  `user_id` varchar(256),
  `is_claimed` boolean,
  PRIMARY KEY (`id`),
  INDEX `idx_lottery_winners_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_lottery_winners_lottery_prize` FOREIGN KEY (`lottery_prize_id`) REFERENCES `lottery_prizes`(`id`),
  CONSTRAINT `fk_lottery_winners_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);

ALTER TABLE `pay_rewards` 
    ADD COLUMN IF NOT EXISTS `lottery_winner_id` varchar(256) NULL;
ALTER TABLE `pay_rewards` 
    ADD CONSTRAINT `fk_pay_rewards_lottery_winner`
    FOREIGN KEY (`lottery_winner_id`) REFERENCES `lottery_winners`(`id`);
