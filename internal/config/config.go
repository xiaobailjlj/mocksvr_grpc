package config

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Management ManagementConfig `mapstructure:"management"`
	MockHTTP   MockHTTPConfig   `mapstructure:"mockhttp"`
	MockGRPC   MockGRPCConfig   `mapstructure:"mockgrpc"`
}

// ServerConfig contains general server settings
type ServerConfig struct {
	RunMode string `mapstructure:"runmode"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// ManagementConfig contains stub management server settings
type ManagementConfig struct {
	Port int `mapstructure:"port"`
}

// MockHTTPConfig contains HTTP mock server settings
type MockHTTPConfig struct {
	Port int `mapstructure:"port"`
}

// MockGRPCConfig contains gRPC mock server settings
type MockGRPCConfig struct {
	Port    int  `mapstructure:"port"`
	Enabled bool `mapstructure:"enabled"`
}
