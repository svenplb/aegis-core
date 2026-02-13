package config

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Scanner: ScannerConfig{
			CustomPatterns: nil,
			Allowlist:      nil,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}
