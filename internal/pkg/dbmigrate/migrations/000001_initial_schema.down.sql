-- Drop all tables in reverse dependency order
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `user_notify_configs`;
DROP TABLE IF EXISTS `alert_channels`;
DROP TABLE IF EXISTS `biz_group_members`;
DROP TABLE IF EXISTS `biz_groups`;
DROP TABLE IF EXISTS `subscribe_rules`;
DROP TABLE IF EXISTS `message_templates`;
DROP TABLE IF EXISTS `notify_medias`;
DROP TABLE IF EXISTS `notify_rules`;
DROP TABLE IF EXISTS `mute_rules`;
DROP TABLE IF EXISTS `notify_records`;
DROP TABLE IF EXISTS `notify_policies`;
DROP TABLE IF EXISTS `notify_channels`;
DROP TABLE IF EXISTS `escalation_steps`;
DROP TABLE IF EXISTS `escalation_policies`;
DROP TABLE IF EXISTS `oncall_shifts`;
DROP TABLE IF EXISTS `schedule_overrides`;
DROP TABLE IF EXISTS `schedule_participants`;
DROP TABLE IF EXISTS `schedules`;
DROP TABLE IF EXISTS `alert_timelines`;
DROP TABLE IF EXISTS `alert_events`;
DROP TABLE IF EXISTS `alert_rule_histories`;
DROP TABLE IF EXISTS `alert_rules`;
DROP TABLE IF EXISTS `datasources`;
DROP TABLE IF EXISTS `team_members`;
DROP TABLE IF EXISTS `teams`;
DROP TABLE IF EXISTS `users`;

SET FOREIGN_KEY_CHECKS = 1;
