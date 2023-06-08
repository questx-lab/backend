CREATE TABLE IF NOT EXISTS `oauth2` (
    `user_id` text, 
    `service` text, 
    `service_user_id` text UNIQUE, 
    PRIMARY KEY (`user_id`, `service`), 
    CONSTRAINT `fk_oauth2_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);
