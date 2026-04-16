ALTER TABLE alert_events ADD COLUMN lark_message_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Lark Bot API message ID for card update (empty when sent via Incoming Webhook)';
