package internal

const (
	EmptyString = ""
)

// EnvKey environment variables
type EnvKey string

const (
	DndUtilRateLimitRps       EnvKey = "DND_UTIL_RATE_LIMIT_RPS"
	DndUtilDbPath             EnvKey = "DND_UTIL_DB_PATH"
	DndUtilTgApiKey           EnvKey = "DND_UTIL_TG_API_KEY"
	DndUtilLongPollingTimeout EnvKey = "DND_UTIL_LONG_POLLING_TIMEOUT"
	DndUtilBotName            EnvKey = "DND_UTIL_BOT_NAME"
)
