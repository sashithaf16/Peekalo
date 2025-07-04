package config

type Config struct {
	LogLevel string // Log level for the application (e.g., "debug", "info", "warn", "error")
	CacheTTL int    // Cache TTL in seconds
}

func GetConfig() *Config {
	return &Config{
		LogLevel: "debug",
		CacheTTL: 60,
	}
}
