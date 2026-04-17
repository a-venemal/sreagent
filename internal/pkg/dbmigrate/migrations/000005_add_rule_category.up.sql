ALTER TABLE `alert_rules` ADD COLUMN `category` varchar(64) NOT NULL DEFAULT '' AFTER `group_name`
