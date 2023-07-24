DROP TABLE IF EXISTS `blockchain_transactions` CASCADE;
DROP TABLE IF EXISTS `pay_rewards` CASCADE;

ALTER TABLE `communities` ADD COLUMN IF NOT EXISTS `wallet_nonce` varchar(256);
ALTER TABLE `claimed_quests` ADD COLUMN IF NOT EXISTS `wallet_address` varchar(256) NULL;

CREATE TABLE IF NOT EXISTS `blockchains` (
  `name` varchar(256),
  `id` bigint UNIQUE,
  `use_external_rpc` boolean,
  `use_eip1559` boolean,
  `block_time` bigint,
  `adjust_time` bigint,
  `threshold_update_block` bigint,
  PRIMARY KEY (`name`)
);

CREATE TABLE IF NOT EXISTS `blockchain_tokens` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `name` varchar(256),
  `chain` varchar(256),
  `symbol` varchar(256),
  `address` varchar(256),
  `decimals` bigint,
  PRIMARY KEY (`id`),
  INDEX `idx_blockchain_tokens_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_blockchain_tokens_chain_token` (`chain`, `address`),
  CONSTRAINT `fk_blockchain_tokens_blockchain` FOREIGN KEY (`chain`) REFERENCES `blockchains`(`name`)
);

CREATE TABLE IF NOT EXISTS `blockchain_connections` (
  `chain` varchar(256),
  `type` varchar(256),
  `url` varchar(256),
  PRIMARY KEY (`chain`, `url`),
  CONSTRAINT `fk_blockchains_blockchain_connections` FOREIGN KEY (`chain`) REFERENCES `blockchains`(`name`)
);

CREATE TABLE IF NOT EXISTS `blockchain_transactions` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `chain` varchar(256),
  `tx_hash` varchar(256),
  `status` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_blockchain_transactions_deleted_at` (`deleted_at`),
  UNIQUE INDEX `idx_blockchain_transaction_chain_txhash` (`chain`, `tx_hash`),
  CONSTRAINT `fk_blockchain_transactions_blockchain` FOREIGN KEY (`chain`) REFERENCES `blockchains`(`name`)
);

CREATE TABLE IF NOT EXISTS `pay_rewards` (
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `token_id` varchar(256),
  `transaction_id` varchar(256),
  `from_community_id` varchar(256),
  `to_user_id` varchar(256),
  `to_address` varchar(256),
  `amount` double,
  `claimed_quest_id` varchar(256),
  `referral_community_id` varchar(256),
  PRIMARY KEY (`id`),
  INDEX `idx_pay_rewards_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_pay_rewards_referral_community` FOREIGN KEY (`referral_community_id`) REFERENCES `communities`(`id`),
  CONSTRAINT `fk_pay_rewards_blockchain_token` FOREIGN KEY (`token_id`) REFERENCES `blockchain_tokens`(`id`),
  CONSTRAINT `fk_pay_rewards_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `blockchain_transactions`(`id`),
  CONSTRAINT `fk_pay_rewards_from_community` FOREIGN KEY (`from_community_id`) REFERENCES `communities`(`id`),
  CONSTRAINT `fk_pay_rewards_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_pay_rewards_claimed_quest` FOREIGN KEY (`claimed_quest_id`) REFERENCES `claimed_quests`(`id`)
);
