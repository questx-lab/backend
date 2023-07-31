ALTER TABLE `chat_channels`
  ADD COLUMN IF NOT EXISTS `description` varchar(256);

