package server

import (
	cmd "github.com/xiaobailjlj/mocksvr_grpc/cmd/root"
	"net"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
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

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func startStubManagementServer(stubHandler *handler.StubHandler, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/url/new", stubHandler.CreateStub)
	mux.HandleFunc("/v1/url/delete", stubHandler.DeleteStub)
	mux.HandleFunc("/v1/url/query/all", stubHandler.GetAllStubs)
	mux.HandleFunc("/v1/url/query/rule", stubHandler.GetRules)

	addr := ":" + strconv.Itoa(port)
	logger.Info("Starting stub management server", zap.String("port", addr))

	if err := http.ListenAndServe(addr, enableCORS(mux)); err != nil {
		logger.Fatal("Failed to start stub management server", zap.Error(err))
	}
}

func startHTTPMockServer(httpHandler *handler.HTTPHandler, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler.ServeMock)

	addr := ":" + strconv.Itoa(port)
	logger.Info("Starting HTTP mock server", zap.String("port", addr))

	if err := http.ListenAndServe(addr, enableCORS(mux)); err != nil {
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
