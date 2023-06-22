CREATE TABLE IF NOT EXISTS `game_luckybox_events` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `room_id` varchar(256),
  `amount` bigint,
  `point_per_box` bigint,
  `is_random` boolean,
  `start_time` datetime NULL,
  `end_time` datetime NULL,
  `is_started` boolean,
  `is_stopped` boolean,
  PRIMARY KEY (`id`),
  INDEX `idx_game_luckybox_events_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_game_luckybox_events_room` FOREIGN KEY (`room_id`) REFERENCES `game_rooms`(`id`)
);

CREATE TABLE IF NOT EXISTS `game_luckyboxes` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `event_id` varchar(256),
  `position_x` bigint,
  `position_y` bigint,
  `point` bigint,
  `is_random` boolean,
  `collected_by` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_game_luckyboxes_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_game_luckyboxes_event` FOREIGN KEY (`event_id`) REFERENCES `game_luckybox_events`(`id`),
  CONSTRAINT `fk_game_luckyboxes_collected_by_user` FOREIGN KEY (`collected_by`) REFERENCES `users`(`id`)
);
