CREATE TABLE `pending_communities` (
    `id` varchar(256), 
    `created_at` datetime NULL, 
    `updated_at` datetime NULL, 
    `deleted_at` datetime NULL, 
    `created_by` varchar(256), 
    `referred_by` varchar(256), 
    `handle` varchar(256), 
    `display_name` varchar(256), 
    `logo_picture` varchar(256), 
    `introduction` longtext, 
    `twitter` varchar(256), 
    `website_url` varchar(256), 
    PRIMARY KEY (`id`), 
    INDEX `idx_pending_communities_deleted_at` (`deleted_at`), 
    CONSTRAINT `fk_pending_communities_created_by_user` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`), 
    CONSTRAINT `fk_pending_communities_referred_by_user` FOREIGN KEY (`referred_by`) REFERENCES `users`(`id`)
);
