package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/config"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"go.uber.org/zap"
)

var (
	cfgFile string
	cfg     config.Config
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mocksvr",
	Short: "Mock server for HTTP and gRPC services",
	Long: `A mock server application that can simulate HTTP and gRPC services 
for testing purposes. It includes management endpoints to configure
the mocked behaviors.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/config.yaml)")
	RootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")

	// Bind flags to viper
	viper.BindPFlag("server.debug", RootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in default locations
		viper.AddConfigPath("./config") // look for config in the config directory
		viper.AddConfigPath(".")        // also look in the working directory
		viper.SetConfigName("config")   // config file name without extension
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// Check if the error is "config file not found"
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create default config file
			defaultConfigPath := filepath.Join("config", "config.yaml")
			os.MkdirAll(filepath.Dir(defaultConfigPath), 0755)
			defaultConfig := `# Server Configuration
server:
  debug: false
  
# Database Configuration
database:
  dsn: "mocksvr:lujing00@tcp(localhost:3306)/mocksvr"

# HTTP Management Server Configuration
management:
  port: 7001
  
# HTTP Mock Server Configuration
mockhttp:
  port: 7002
  
# gRPC Mock Server Configuration
mockgrpc:
  port: 7003
  enabled: false
`
			if err := os.WriteFile(defaultConfigPath, []byte(defaultConfig), 0644); err == nil {
				fmt.Println("Created default config file at:", defaultConfigPath)
				viper.SetConfigFile(defaultConfigPath)
				if err := viper.ReadInConfig(); err != nil {
					fmt.Println("Warning: Could not read created config file:", err)
				}
			} else {
				fmt.Println("Warning: Could not create default config file:", err)
			}
		} else {
			fmt.Println("Warning: Error reading config file:", err)
		}
	}

	// Unmarshal config into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Unable to decode config into struct:", err)
	}

	// Initialize logger based on config
	logger.InitLogger(cfg.Server.Debug)
	logger.Info("Configuration loaded successfully",
		zap.String("config_file", viper.ConfigFileUsed()),
		zap.Bool("debug", cfg.Server.Debug))
}

// GetConfig returns the current configuration
func GetConfig() config.Config {
	return cfg
}
