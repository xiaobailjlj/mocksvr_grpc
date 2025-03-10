package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/config"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"go.uber.org/zap"
	"os"
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
		fmt.Println("Error reading config file:", err)
	}

	// Unmarshal config into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Unable to decode config into struct:", err)
	}

	// Initialize logger based on config
	logger.InitLogger(cfg.Server.RunMode)
	logger.Info("Configuration loaded successfully",
		zap.String("config_file", viper.ConfigFileUsed()),
		zap.String("run_mode", cfg.Server.RunMode))
}

// GetConfig returns the current configuration
func GetConfig() config.Config {
	return cfg
}
