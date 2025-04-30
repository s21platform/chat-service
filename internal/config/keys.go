package config

type key string

const (
	KeyUUID                         = key("uuid")
	KeyLogger                       = key("logger")
	KeyMetrics                  key = key("metrics")
	UserNicknameConsumerGroupID     = "chat-nickname-updater"
)
