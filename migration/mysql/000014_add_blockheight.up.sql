ALTER TABLE `blockchain_transactions`
  ADD COLUMN IF NOT EXISTS `block_height` bigint;

CREATE INDEX idx_blockchain_transactions_tx_hash_chain_id ON blockchain_transactions(tx_hash, chain);

