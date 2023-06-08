CREATE TABLE IF NOT EXISTS `oauth2` (
    `user_id` varchar(256), 
    `service` varchar(256), 
    `service_user_id` varchar(256) UNIQUE, 
    PRIMARY KEY (`user_id`, `service`), 
    CONSTRAINT `fk_oauth2_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);
