ALTER TABLE `game_rooms`
    ADD `community_id` varchar(256);

ALTER TABLE `game_rooms`
    ADD CONSTRAINT `fk_game_rooms_community_id`
    FOREIGN KEY(`community_id`)
    REFERENCES `communities`(`id`);
