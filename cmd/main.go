package main

import (
	cmd "github.com/xiaobailjlj/mocksvr_grpc/cmd/root"
	"github.com/xiaobailjlj/mocksvr_grpc/cmd/server"
	"github.com/xiaobailjlj/mocksvr_grpc/cmd/version"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
)

func main() {
	// Add commands to root command
	cmd.RootCmd.AddCommand(server.NewServerCmd())
	cmd.RootCmd.AddCommand(version.NewVersionCmd())

	// Execute the root command
	cmd.Execute()

	// Ensure logs are flushed
	defer logger.Logger.Sync()
}
