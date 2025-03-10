package server

import (
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	cmd "github.com/xiaobailjlj/mocksvr_grpc/cmd/root"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/handler"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewServerCmd creates a new server command
func NewServerCmd() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the mock server",
		Long:  `Start the mock server with management, HTTP, and optionally gRPC services.`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}
	return serverCmd
}

func runServer() {
	cfg := cmd.GetConfig()
	logger.Info("Starting mock server application")

	// Set Gin mode based on config
	gin.SetMode(getGinMode(cfg.Server.RunMode))

	// Initialize database connection
	mysqlStorage, err := storage.NewMySQLStorage(cfg.Database.DSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer mysqlStorage.Close()
	logger.Info("Successfully connected to database")

	// Initialize services and handlers
	mockService := service.NewMockService(mysqlStorage)
	stubHandler := handler.NewStubHandler(mockService)
	httpHandler := handler.NewHTTPHandler(mockService)

	// Start servers
	go startStubManagementServer(stubHandler, cfg.Management.Port)
	go startHTTPMockServer(httpHandler, cfg.MockHTTP.Port)

	// Start gRPC server if enabled
	if cfg.MockGRPC.Enabled {
		go startGRPCMockServer(mockService, cfg.MockGRPC.Port)
	} else {
		logger.Info("gRPC mock server is disabled")
	}

	logger.Info("All servers started successfully")
	select {} // Block forever
}

// MyBenchLogger is a middleware for benchmark endpoint logging
func MyBenchLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Benchmark endpoint called",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()))
		c.Next()
	}
}

// Get Gin mode based on the application run mode
func getGinMode(runMode string) string {
	logger.Info("Setting Gin mode", zap.String("runMode", runMode))
	switch runMode {
	case "debug":
		return gin.DebugMode
	case "test":
		return gin.TestMode
	default:
		return gin.ReleaseMode
	}
}

// CORS middleware for Gin
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func startStubManagementServer(stubHandler *handler.StubHandler, port int) {
	r := gin.New()

	// Apply middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Define routes
	v1 := r.Group("/v1/url")
	{
		v1.POST("/new", func(c *gin.Context) {
			stubHandler.CreateStubGin(c)
		})
		v1.DELETE("/delete", func(c *gin.Context) {
			stubHandler.DeleteStubGin(c)
		})
		v1.GET("/query/all", func(c *gin.Context) {
			stubHandler.GetAllStubsGin(c)
		})
		v1.GET("/query/rule", func(c *gin.Context) {
			stubHandler.GetRulesGin(c)
		})
	}

	// Add benchmark endpoint
	r.GET("/benchmark", MyBenchLogger(), benchEndpoint)

	addr := ":" + strconv.Itoa(port)
	logger.Info("Starting stub management server with Gin", zap.String("port", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start stub management server", zap.Error(err))
	}
}

func benchEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Benchmark endpoint reached",
	})
}

func startHTTPMockServer(httpHandler *handler.HTTPHandler, port int) {
	r := gin.New()

	// Apply middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Catch all route
	r.Any("/*path", func(c *gin.Context) {
		httpHandler.ServeMockGin(c)
	})

	addr := ":" + strconv.Itoa(port)
	logger.Info("Starting HTTP mock server with Gin", zap.String("port", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start HTTP mock server", zap.Error(err))
	}
}

func startGRPCMockServer(mockService *service.MockService, port int) {
	addr := ":" + strconv.Itoa(port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()
	handler.RegisterGRPCServer(s, mockService)
	logger.Info("Starting gRPC mock server", zap.String("port", addr))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}
