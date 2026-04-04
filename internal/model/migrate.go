package model

// NotificationV2Models returns all new notification system v2 models
// that need to be auto-migrated. This function is called by main.go
// during database initialization.
func NotificationV2Models() []interface{} {
	return []interface{}{
		&NotifyRule{},
		&NotifyMedia{},
		&MessageTemplate{},
		&SubscribeRule{},
		&BizGroup{},
		&BizGroupMember{},
	}
}

// DispatchModels returns models for the alert channel and user notify config system.
func DispatchModels() []interface{} {
	return []interface{}{
		&AlertChannel{},
		&UserNotifyConfig{},
	}
}
