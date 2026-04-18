ALTER TABLE alert_rules
  ADD COLUMN rule_type VARCHAR(32) NOT NULL DEFAULT 'threshold' AFTER name,
  ADD COLUMN heartbeat_token VARCHAR(128) NOT NULL DEFAULT '' AFTER rule_type,
  ADD COLUMN heartbeat_interval INT NOT NULL DEFAULT 300 AFTER heartbeat_token,
  ADD COLUMN heartbeat_last_at DATETIME(3) NULL AFTER heartbeat_interval,
  ADD COLUMN ack_sla_minutes INT NOT NULL DEFAULT 0 AFTER heartbeat_last_at
