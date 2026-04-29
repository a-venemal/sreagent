CREATE TABLE IF NOT EXISTS event_pipelines (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description VARCHAR(1024) DEFAULT '',
    disabled TINYINT(1) DEFAULT 0,
    filter_enable TINYINT(1) DEFAULT 0,
    label_filters JSON,
    nodes JSON,
    connections JSON,
    created_by BIGINT UNSIGNED DEFAULT 0,
    updated_by BIGINT UNSIGNED DEFAULT 0,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pipeline_executions (
    id VARCHAR(36) PRIMARY KEY,
    pipeline_id BIGINT UNSIGNED NOT NULL,
    event_id BIGINT UNSIGNED DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    node_results JSON,
    error_message TEXT,
    duration_ms BIGINT DEFAULT 0,
    started_at DATETIME(3) NOT NULL,
    finished_at DATETIME(3) NOT NULL,
    INDEX idx_pipeline_id (pipeline_id),
    INDEX idx_event_id (event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
