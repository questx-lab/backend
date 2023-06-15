ALTER TABLE `pay_rewards`
  DROP FOREIGN KEY `fk_pay_rewards_claimed_quest`;

ALTER TABLE `pay_rewards`
  DROP FOREIGN KEY `fk_pay_rewards_user`;

ALTER TABLE `pay_rewards`
  DROP COLUMN IF EXISTS `tx_hash`;

ALTER TABLE `pay_rewards`
  DROP COLUMN IF EXISTS `user_id`;

ALTER TABLE `pay_rewards`
  DROP COLUMN IF EXISTS `address`;

ALTER TABLE `pay_rewards`
  DROP COLUMN IF EXISTS `claimed_quest_id`;

ALTER TABLE `pay_rewards`
  ADD COLUMN IF NOT EXISTS `to_address` varchar(256);

ALTER TABLE `pay_rewards`
  ADD COLUMN IF NOT EXISTS `to_user_id` varchar(256);

ALTER TABLE `pay_rewards`
  ADD COLUMN IF NOT EXISTS `is_received` boolean;

CREATE TABLE IF NOT EXISTS `blockchain_transactions`(
  `id` varchar(256),
  `created_at` datetime NULL,
  `updated_at` datetime NULL,
  `deleted_at` datetime NULL,
  `pay_reward_id` varchar(256),
  `chain` varchar(256),
  `status` varchar(256),
  `address` varchar(256),
  `token` varchar(256),
  `amount` double,
  `tx_hash` varchar(256),
  `tx_bytes` longtext,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_blockchain_transactions_pay_reward` FOREIGN KEY (`pay_reward_id`) REFERENCES `pay_rewards`(`id`),
);

