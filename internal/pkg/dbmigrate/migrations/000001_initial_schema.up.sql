-- =============================================================================
-- SREAgent Initial Schema  v1
-- 适用于 MySQL 8.0+
-- 所有表使用 utf8mb4_unicode_ci，软删除字段 deleted_at，JSON 列用 json 类型
-- =============================================================================

-- -----------------------------------------------------------------------------
-- users
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id`             bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`     datetime(3)      NULL,
  `updated_at`     datetime(3)      NULL,
  `deleted_at`     datetime(3)      NULL,
  `username`       varchar(64)      NOT NULL,
  `password`       varchar(256)     NOT NULL,
  `display_name`   varchar(128)     DEFAULT NULL,
  `email`          varchar(256)     DEFAULT NULL,
  `phone`          varchar(32)      DEFAULT NULL,
  `lark_user_id`   varchar(64)      DEFAULT NULL,
  `avatar`         varchar(512)     DEFAULT NULL,
  `role`           varchar(32)      NOT NULL DEFAULT 'member',
  `is_active`      tinyint(1)       DEFAULT 1,
  `user_type`      varchar(32)      DEFAULT 'human',
  `notify_target`  text             DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username`    (`username`),
  KEY               `idx_users_deleted_at`  (`deleted_at`),
  KEY               `idx_users_lark_user_id`(`lark_user_id`),
  KEY               `idx_users_user_type`   (`user_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- teams
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `teams` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `description` varchar(512)     DEFAULT NULL,
  `labels`      json             DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_teams_name`       (`name`),
  KEY            `idx_teams_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- team_members  (many2many join table)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `team_members` (
  `team_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `role`    varchar(32)     NOT NULL DEFAULT 'member',
  PRIMARY KEY (`team_id`, `user_id`),
  KEY `idx_team_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- datasources
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `datasources` (
  `id`                    bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`            datetime(3)      NULL,
  `updated_at`            datetime(3)      NULL,
  `deleted_at`            datetime(3)      NULL,
  `name`                  varchar(128)     NOT NULL,
  `type`                  varchar(32)      NOT NULL,
  `endpoint`              varchar(512)     NOT NULL,
  `description`           varchar(512)     DEFAULT NULL,
  `labels`                json             DEFAULT NULL,
  `status`                varchar(32)      DEFAULT 'unknown',
  `auth_type`             varchar(32)      DEFAULT NULL,
  `auth_config`           text             DEFAULT NULL,
  `health_check_interval` int              DEFAULT 60,
  `is_enabled`            tinyint(1)       DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_datasources_name`       (`name`),
  KEY            `idx_datasources_deleted_at` (`deleted_at`),
  KEY            `idx_datasources_type`       (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- alert_rules
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alert_rules` (
  `id`               bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`       datetime(3)      NULL,
  `updated_at`       datetime(3)      NULL,
  `deleted_at`       datetime(3)      NULL,
  `name`             varchar(256)     NOT NULL,
  `display_name`     varchar(256)     DEFAULT NULL,
  `description`      text             DEFAULT NULL,
  `data_source_id`   bigint unsigned  NOT NULL DEFAULT 0,
  `expression`       text             NOT NULL,
  `for_duration`     varchar(32)      DEFAULT '0s',
  `severity`         varchar(32)      NOT NULL,
  `labels`           json             DEFAULT NULL,
  `annotations`      json             DEFAULT NULL,
  `status`           varchar(32)      DEFAULT 'enabled',
  `group_name`       varchar(128)     DEFAULT NULL,
  `version`          int              DEFAULT 1,
  `created_by`       bigint unsigned  DEFAULT NULL,
  `updated_by`       bigint unsigned  DEFAULT NULL,
  `eval_interval`    int              DEFAULT 60,
  `recovery_hold`    varchar(32)      DEFAULT '0s',
  `no_data_enabled`  tinyint(1)       DEFAULT 0,
  `no_data_duration` varchar(32)      DEFAULT '5m',
  `suppress_enabled` tinyint(1)       DEFAULT 0,
  `biz_group_id`     bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_rules_deleted_at`    (`deleted_at`),
  KEY `idx_alert_rules_name`          (`name`),
  KEY `idx_alert_rules_data_source_id`(`data_source_id`),
  KEY `idx_alert_rules_severity`      (`severity`),
  KEY `idx_alert_rules_status`        (`status`),
  KEY `idx_alert_rules_group_name`    (`group_name`),
  KEY `idx_alert_rules_created_by`    (`created_by`),
  KEY `idx_alert_rules_biz_group_id`  (`biz_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- alert_rule_histories
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alert_rule_histories` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `rule_id`     bigint unsigned  NOT NULL,
  `version`     int              NOT NULL,
  `change_type` varchar(32)      NOT NULL,
  `snapshot`    text             NOT NULL,
  `changed_by`  bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_rule_histories_deleted_at` (`deleted_at`),
  KEY `idx_alert_rule_histories_rule_id`    (`rule_id`),
  KEY `idx_alert_rule_histories_changed_by` (`changed_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- alert_events
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alert_events` (
  `id`             bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`     datetime(3)      NULL,
  `updated_at`     datetime(3)      NULL,
  `deleted_at`     datetime(3)      NULL,
  `fingerprint`    varchar(64)      NOT NULL,
  `rule_id`        bigint unsigned  DEFAULT NULL,
  `alert_name`     varchar(256)     NOT NULL,
  `severity`       varchar(32)      NOT NULL,
  `status`         varchar(32)      NOT NULL DEFAULT 'firing',
  `labels`         json             DEFAULT NULL,
  `annotations`    json             DEFAULT NULL,
  `source`         varchar(128)     DEFAULT NULL,
  `generator_url`  varchar(512)     DEFAULT NULL,
  `fired_at`       datetime(3)      NOT NULL,
  `acked_at`       datetime(3)      DEFAULT NULL,
  `resolved_at`    datetime(3)      DEFAULT NULL,
  `closed_at`      datetime(3)      DEFAULT NULL,
  `acked_by`       bigint unsigned  DEFAULT NULL,
  `assigned_to`    bigint unsigned  DEFAULT NULL,
  `silenced_until` datetime(3)      DEFAULT NULL,
  `silence_reason` varchar(512)     DEFAULT NULL,
  `resolution`     text             DEFAULT NULL,
  `fire_count`     int              DEFAULT 1,
  `on_call_user_id`bigint unsigned  DEFAULT NULL,
  `is_dispatched`  tinyint(1)       DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_alert_events_fingerprint` (`fingerprint`),
  KEY `idx_alert_events_deleted_at`    (`deleted_at`),
  KEY `idx_alert_events_rule_id`       (`rule_id`),
  KEY `idx_alert_events_alert_name`    (`alert_name`),
  KEY `idx_alert_events_severity`      (`severity`),
  KEY `idx_alert_events_status`        (`status`),
  KEY `idx_alert_events_fired_at`      (`fired_at`),
  KEY `idx_alert_events_acked_by`      (`acked_by`),
  KEY `idx_alert_events_assigned_to`   (`assigned_to`),
  KEY `idx_alert_events_silenced_until`(`silenced_until`),
  KEY `idx_alert_events_on_call_user_id`(`on_call_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- alert_timelines
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alert_timelines` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `event_id`    bigint unsigned  NOT NULL,
  `action`      varchar(32)      NOT NULL,
  `operator_id` bigint unsigned  DEFAULT NULL,
  `note`        text             DEFAULT NULL,
  `extra`       json             DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_timelines_deleted_at`  (`deleted_at`),
  KEY `idx_alert_timelines_event_id`    (`event_id`),
  KEY `idx_alert_timelines_operator_id` (`operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- schedules
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `schedules` (
  `id`              bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`      datetime(3)      NULL,
  `updated_at`      datetime(3)      NULL,
  `deleted_at`      datetime(3)      NULL,
  `name`            varchar(128)     NOT NULL,
  `team_id`         bigint unsigned  DEFAULT NULL,
  `description`     varchar(512)     DEFAULT NULL,
  `rotation_type`   varchar(32)      NOT NULL,
  `timezone`        varchar(64)      DEFAULT 'Asia/Shanghai',
  `handoff_time`    varchar(8)       DEFAULT '09:00',
  `handoff_day`     int              DEFAULT 1,
  `is_enabled`      tinyint(1)       DEFAULT 1,
  `severity_filter` varchar(128)     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_schedules_deleted_at` (`deleted_at`),
  KEY `idx_schedules_team_id`    (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- schedule_participants
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `schedule_participants` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `schedule_id` bigint unsigned  NOT NULL,
  `user_id`     bigint unsigned  NOT NULL,
  `position`    int              NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_schedule_participants_deleted_at`  (`deleted_at`),
  KEY `idx_schedule_participants_schedule_id` (`schedule_id`),
  KEY `idx_schedule_participants_user_id`     (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- schedule_overrides
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `schedule_overrides` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `schedule_id` bigint unsigned  NOT NULL,
  `user_id`     bigint unsigned  NOT NULL,
  `start_time`  datetime(3)      NOT NULL,
  `end_time`    datetime(3)      NOT NULL,
  `reason`      varchar(256)     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_schedule_overrides_deleted_at`  (`deleted_at`),
  KEY `idx_schedule_overrides_schedule_id` (`schedule_id`),
  KEY `idx_schedule_overrides_user_id`     (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- oncall_shifts
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `oncall_shifts` (
  `id`              bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`      datetime(3)      NULL,
  `updated_at`      datetime(3)      NULL,
  `deleted_at`      datetime(3)      NULL,
  `schedule_id`     bigint unsigned  NOT NULL,
  `user_id`         bigint unsigned  NOT NULL,
  `start_time`      datetime(3)      NOT NULL,
  `end_time`        datetime(3)      NOT NULL,
  `severity_filter` varchar(128)     DEFAULT NULL,
  `source`          varchar(32)      DEFAULT 'manual',
  `note`            varchar(256)     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_oncall_shifts_deleted_at`  (`deleted_at`),
  KEY `idx_oncall_shifts_schedule_id` (`schedule_id`),
  KEY `idx_oncall_shifts_user_id`     (`user_id`),
  KEY `idx_oncall_shifts_start_time`  (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- escalation_policies
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `escalation_policies` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `team_id`     bigint unsigned  NOT NULL DEFAULT 0,
  `is_enabled`  tinyint(1)       DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_escalation_policies_deleted_at` (`deleted_at`),
  KEY `idx_escalation_policies_team_id`    (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- escalation_steps
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `escalation_steps` (
  `id`                bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`        datetime(3)      NULL,
  `updated_at`        datetime(3)      NULL,
  `deleted_at`        datetime(3)      NULL,
  `policy_id`         bigint unsigned  NOT NULL,
  `step_order`        int              NOT NULL,
  `delay_minutes`     int              NOT NULL,
  `target_type`       varchar(32)      NOT NULL,
  `target_id`         bigint unsigned  NOT NULL DEFAULT 0,
  `notify_channel_id` bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_escalation_steps_deleted_at`        (`deleted_at`),
  KEY `idx_escalation_steps_policy_id`         (`policy_id`),
  KEY `idx_escalation_steps_notify_channel_id` (`notify_channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- notify_channels
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `notify_channels` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `type`        varchar(32)      NOT NULL,
  `description` varchar(512)     DEFAULT NULL,
  `labels`      json             DEFAULT NULL,
  `config`      text             NOT NULL,
  `is_enabled`  tinyint(1)       DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_notify_channels_deleted_at` (`deleted_at`),
  KEY `idx_notify_channels_type`       (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- notify_policies
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `notify_policies` (
  `id`               bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`       datetime(3)      NULL,
  `updated_at`       datetime(3)      NULL,
  `deleted_at`       datetime(3)      NULL,
  `name`             varchar(128)     NOT NULL,
  `description`      varchar(512)     DEFAULT NULL,
  `match_labels`     json             DEFAULT NULL,
  `severities`       varchar(128)     DEFAULT NULL,
  `channel_id`       bigint unsigned  NOT NULL DEFAULT 0,
  `throttle_minutes` int              DEFAULT 5,
  `template_name`    varchar(64)      DEFAULT 'default',
  `is_enabled`       tinyint(1)       DEFAULT 1,
  `priority`         int              DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_notify_policies_deleted_at` (`deleted_at`),
  KEY `idx_notify_policies_channel_id` (`channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- notify_records
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `notify_records` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `event_id`    bigint unsigned  NOT NULL,
  `channel_id`  bigint unsigned  NOT NULL DEFAULT 0,
  `policy_id`   bigint unsigned  DEFAULT 0,
  `status`      varchar(32)      NOT NULL,
  `response`    text             DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notify_records_deleted_at` (`deleted_at`),
  KEY `idx_notify_records_event_id`   (`event_id`),
  KEY `idx_notify_records_channel_id` (`channel_id`),
  KEY `idx_notify_records_policy_id`  (`policy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- mute_rules
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `mute_rules` (
  `id`              bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`      datetime(3)      NULL,
  `updated_at`      datetime(3)      NULL,
  `deleted_at`      datetime(3)      NULL,
  `name`            varchar(128)     NOT NULL,
  `description`     varchar(512)     DEFAULT NULL,
  `match_labels`    json             DEFAULT NULL,
  `severities`      varchar(128)     DEFAULT NULL,
  `start_time`      datetime(3)      DEFAULT NULL,
  `end_time`        datetime(3)      DEFAULT NULL,
  `periodic_start`  varchar(8)       DEFAULT NULL,
  `periodic_end`    varchar(8)       DEFAULT NULL,
  `days_of_week`    varchar(32)      DEFAULT NULL,
  `timezone`        varchar(64)      DEFAULT 'Asia/Shanghai',
  `created_by`      bigint unsigned  DEFAULT NULL,
  `is_enabled`      tinyint(1)       DEFAULT 1,
  `rule_ids`        varchar(512)     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_mute_rules_deleted_at`  (`deleted_at`),
  KEY `idx_mute_rules_created_by`  (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- notify_rules
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `notify_rules` (
  `id`              bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`      datetime(3)      NULL,
  `updated_at`      datetime(3)      NULL,
  `deleted_at`      datetime(3)      NULL,
  `name`            varchar(128)     NOT NULL,
  `description`     varchar(512)     DEFAULT NULL,
  `is_enabled`      tinyint(1)       DEFAULT 1,
  `severities`      varchar(128)     DEFAULT NULL,
  `match_labels`    json             DEFAULT NULL,
  `pipeline`        text             DEFAULT NULL,
  `notify_configs`  text             DEFAULT NULL,
  `repeat_interval` int              DEFAULT 3600,
  `callback_url`    varchar(512)     DEFAULT NULL,
  `created_by`      bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notify_rules_deleted_at`  (`deleted_at`),
  KEY `idx_notify_rules_created_by`  (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- notify_medias
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `notify_medias` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `type`        varchar(32)      NOT NULL,
  `description` varchar(512)     DEFAULT NULL,
  `is_enabled`  tinyint(1)       DEFAULT 1,
  `config`      text             NOT NULL,
  `variables`   text             DEFAULT NULL,
  `is_builtin`  tinyint(1)       DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_notify_medias_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- message_templates
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `message_templates` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `description` varchar(512)     DEFAULT NULL,
  `content`     text             NOT NULL,
  `type`        varchar(32)      DEFAULT 'text',
  `is_builtin`  tinyint(1)       DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_message_templates_name`       (`name`),
  KEY            `idx_message_templates_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- subscribe_rules
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `subscribe_rules` (
  `id`             bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`     datetime(3)      NULL,
  `updated_at`     datetime(3)      NULL,
  `deleted_at`     datetime(3)      NULL,
  `name`           varchar(128)     NOT NULL,
  `description`    varchar(512)     DEFAULT NULL,
  `is_enabled`     tinyint(1)       DEFAULT 1,
  `match_labels`   json             DEFAULT NULL,
  `severities`     varchar(128)     DEFAULT NULL,
  `notify_rule_id` bigint unsigned  NOT NULL DEFAULT 0,
  `user_id`        bigint unsigned  DEFAULT NULL,
  `team_id`        bigint unsigned  DEFAULT NULL,
  `created_by`     bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_subscribe_rules_deleted_at`    (`deleted_at`),
  KEY `idx_subscribe_rules_notify_rule_id`(`notify_rule_id`),
  KEY `idx_subscribe_rules_user_id`       (`user_id`),
  KEY `idx_subscribe_rules_team_id`       (`team_id`),
  KEY `idx_subscribe_rules_created_by`    (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- biz_groups
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `biz_groups` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `name`        varchar(128)     NOT NULL,
  `description` varchar(512)     DEFAULT NULL,
  `parent_id`   bigint unsigned  DEFAULT NULL,
  `labels`      json             DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_biz_groups_deleted_at` (`deleted_at`),
  KEY `idx_biz_groups_parent_id`  (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- biz_group_members  (join table)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `biz_group_members` (
  `biz_group_id` bigint unsigned NOT NULL,
  `user_id`      bigint unsigned NOT NULL,
  `role`         varchar(32)     NOT NULL DEFAULT 'member',
  PRIMARY KEY (`biz_group_id`, `user_id`),
  KEY `idx_biz_group_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- alert_channels
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `alert_channels` (
  `id`           bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`   datetime(3)      NULL,
  `updated_at`   datetime(3)      NULL,
  `deleted_at`   datetime(3)      NULL,
  `name`         varchar(128)     NOT NULL,
  `description`  varchar(512)     DEFAULT NULL,
  `match_labels` json             DEFAULT NULL,
  `severities`   varchar(128)     DEFAULT NULL,
  `media_id`     bigint unsigned  NOT NULL DEFAULT 0,
  `template_id`  bigint unsigned  DEFAULT NULL,
  `throttle_min` int              DEFAULT 5,
  `is_enabled`   tinyint(1)       DEFAULT 1,
  `created_by`   bigint unsigned  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_alert_channels_deleted_at`  (`deleted_at`),
  KEY `idx_alert_channels_media_id`    (`media_id`),
  KEY `idx_alert_channels_template_id` (`template_id`),
  KEY `idx_alert_channels_created_by`  (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- user_notify_configs
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user_notify_configs` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `created_at`  datetime(3)      NULL,
  `updated_at`  datetime(3)      NULL,
  `deleted_at`  datetime(3)      NULL,
  `user_id`     bigint unsigned  NOT NULL,
  `media_type`  varchar(32)      DEFAULT NULL,
  `config`      text             DEFAULT NULL,
  `is_enabled`  tinyint(1)       DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `udx_user_media`              (`user_id`, `media_type`),
  KEY            `idx_user_notify_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
