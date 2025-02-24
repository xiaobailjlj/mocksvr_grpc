// cmd/server/main.go
package main

import (
	"flag"
	"go.uber.org/zap"
	"net"
	"net/http"

	"github.com/xiaobailjlj/mocksvr_grpc/internal/handler"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	"google.golang.org/grpc"
)

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

func startStubManagementServer(stubHandler *handler.StubHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/url/new", stubHandler.CreateStub)
	mux.HandleFunc("/v1/url/delete", stubHandler.DeleteStub)
	mux.HandleFunc("/v1/url/query/all", stubHandler.GetAllStubs)
	mux.HandleFunc("/v1/url/query/rule", stubHandler.GetRules)

	logger.Info("Starting stub management server", zap.String("port", "7001"))
	if err := http.ListenAndServe(":7001", enableCORS(mux)); err != nil {
		logger.Fatal("Failed to start stub management server", zap.Error(err))
	}
}

func startHTTPMockServer(httpHandler *handler.HTTPHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler.ServeMock)

	logger.Info("Starting HTTP mock server", zap.String("port", "7002"))
	if err := http.ListenAndServe(":7002", enableCORS(mux)); err != nil {
		logger.Fatal("Failed to start HTTP mock server", zap.Error(err))
	}
}

func startGRPCMockServer(mockService *service.MockService) {
	lis, err := net.Listen("tcp", ":7003")
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	handler.RegisterGRPCServer(s, mockService)

	logger.Info("Starting gRPC mock server", zap.String("port", "7003"))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	logger.InitLogger(*debug)
	defer logger.Logger.Sync()

	logger.Info("Starting mock server application")

	mysqlStorage, err := storage.NewMySQLStorage("mocksvr:lujing00@tcp(localhost:3306)/mocksvr")

	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer mysqlStorage.Close()

	logger.Info("Successfully connected to database")

	mockService := service.NewMockService(mysqlStorage)
	stubHandler := handler.NewStubHandler(mockService)
	httpHandler := handler.NewHTTPHandler(mockService)

	go startStubManagementServer(stubHandler)
	go startHTTPMockServer(httpHandler)
	//go startGRPCMockServer(mockService)

	logger.Info("All servers started successfully")

	select {}
}
