ALTER TABLE alert_rules ADD COLUMN group_wait_seconds INT NOT NULL DEFAULT 0 AFTER recovery_hold, ADD COLUMN group_interval_seconds INT NOT NULL DEFAULT 0 AFTER group_wait_seconds;
