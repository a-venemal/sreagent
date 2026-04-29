ALTER TABLE alert_rules
  DROP COLUMN rule_type,
  DROP COLUMN heartbeat_token,
  DROP COLUMN heartbeat_interval,
  DROP COLUMN heartbeat_last_at,
  DROP COLUMN ack_sla_minutes
