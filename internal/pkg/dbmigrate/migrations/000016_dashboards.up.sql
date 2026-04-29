CREATE TABLE IF NOT EXISTS dashboards (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description VARCHAR(1024) DEFAULT '',
    tags JSON,
    config LONGTEXT,
    created_by BIGINT UNSIGNED DEFAULT 0,
    updated_by BIGINT UNSIGNED DEFAULT 0,
    is_public TINYINT(1) DEFAULT 0,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
