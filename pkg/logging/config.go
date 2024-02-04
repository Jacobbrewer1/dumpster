package logging

// Config is the configuration for the logging.
type Config struct {
	// appName is the name of the application.
	appName Name
}

// NewConfig creates a new Config.
func NewConfig(appName Name) *Config {
	return &Config{
		appName: appName,
	}
}
