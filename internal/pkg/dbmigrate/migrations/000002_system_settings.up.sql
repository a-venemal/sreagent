-- 000002_system_settings.up.sql
-- Stores platform-level configuration (AI, Lark Bot, etc.) as key-value pairs.
-- Replaces config.yaml / K8s Secrets for non-startup settings.

CREATE TABLE IF NOT EXISTS system_settings (
    id          INT UNSIGNED    NOT NULL AUTO_INCREMENT,
    `group`     VARCHAR(64)     NOT NULL COMMENT 'setting group, e.g. ai / lark',
    `key`       VARCHAR(128)    NOT NULL COMMENT 'setting key within the group',
    `value`     TEXT            NOT NULL COMMENT 'setting value (plaintext; secrets should be rotated via Web UI)',
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY idx_group_key (`group`, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Platform-level key-value settings (AI, Lark Bot, etc.)';
