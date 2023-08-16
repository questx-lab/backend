CREATE TABLE IF NOT EXISTS `non_fungible_tokens` (
  `id` bigint,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256),
  `created_by` varchar(256),
  `chain` varchar(256),
  `title` varchar(256),
  `image_url` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_non_fungible_tokens_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_non_fungible_tokens_created_by_user` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_non_fungible_tokens_blockchain` FOREIGN KEY (`chain`) REFERENCES `blockchains`(`name`),
  CONSTRAINT `fk_non_fungible_tokens_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`)
);

CREATE TABLE IF NOT EXISTS `non_fungible_token_mint_histories` (
  `non_fungible_token_id` bigint,
  `created_at` datetime NULL,
  `transaction_id` varchar(256),
  `count` bigint,
  PRIMARY KEY (`non_fungible_token_id`),
  CONSTRAINT `fk_non_fungible_token_mint_histories_non_fungible_token` FOREIGN KEY (`non_fungible_token_id`) REFERENCES `non_fungible_tokens`(`id`),
  CONSTRAINT `fk_non_fungible_token_mint_histories_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `blockchain_transactions`(`id`)
);

ALTER TABLE `pay_rewards` ADD IF NOT EXISTS `non_fungible_token_id` bigint NULL;

ALTER TABLE `pay_rewards` ADD CONSTRAINT `fk_pay_rewards_non_fungible_token` 
  FOREIGN KEY (`non_fungible_token_id`) REFERENCES `non_fungible_tokens`(`id`);

ALTER TABLE `blockchains` ADD IF NOT EXISTS `xquest_nft_address` varchar(256);
