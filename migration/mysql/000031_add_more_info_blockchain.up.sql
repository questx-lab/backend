ALTER TABLE `blockchains` ADD IF NOT EXISTS `display_name` VARCHAR(256);
ALTER TABLE `blockchains` ADD IF NOT EXISTS `currency_symbol` VARCHAR(256);
ALTER TABLE `blockchains` ADD IF NOT EXISTS `explorer_url` VARCHAR(256);
