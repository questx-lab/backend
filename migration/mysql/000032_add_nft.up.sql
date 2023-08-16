CREATE TABLE IF NOT EXISTS `nft_sets`(
  `id` bigint,
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `community_id` varchar(256),
  `title` varchar(256),
  `image_url` varchar(256),
  `chain` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_nft_sets_deleted_at`(`deleted_at`),
  CONSTRAINT `fk_nft_sets_community` FOREIGN KEY (`community_id`) REFERENCES `communities`(`id`),
  CONSTRAINT `fk_nft_sets_blockchain` FOREIGN KEY (`chain`) REFERENCES `blockchains`(`name`)
);

CREATE TABLE IF NOT EXISTS `nfts`(
  `id` bigint,
  `set_id` bigint,
  `transaction_id` varchar(256),
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_nfts_set` FOREIGN KEY (`set_id`) REFERENCES `nft_sets`(`id`),
  CONSTRAINT `fk_nfts_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `blockchain_transactions`(`id`)
);

ALTER TABLE `pay_rewards`
  ADD IF NOT EXISTS `nft_id` VARCHAR(256) NULL;

ALTER TABLE `pay_rewards`
  ADD CONSTRAINT `fk_pay_rewards_nft` FOREIGN KEY (`nft_id`) REFERENCES `nfts`(`id`);

