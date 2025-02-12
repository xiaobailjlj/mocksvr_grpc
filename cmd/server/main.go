package main

import (
	"log"
	"net"
	"net/http"

	"github.com/xiaobailjlj/mocksvr_grpc/internal/handler"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	"google.golang.org/grpc"
)

func startStubManagementServer(stubHandler *handler.StubHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/url/new", stubHandler.CreateStub)

	log.Printf("Stub management server listening on :7001")
	if err := http.ListenAndServe(":7001", mux); err != nil {
		log.Fatalf("failed to serve stub management server: %v", err)
	}
}

func startHTTPMockServer(httpHandler *handler.HTTPHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler.ServeMock)

	log.Printf("HTTP mock server listening on :7002")
	if err := http.ListenAndServe(":7002", mux); err != nil {
		log.Fatalf("failed to serve HTTP mock server: %v", err)
	}
}

func startGRPCMockServer(mockService *service.MockService) {
	lis, err := net.Listen("tcp", ":7003")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	handler.RegisterGRPCServer(s, mockService)

	log.Printf("gRPC mock server listening on :7003")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func main() {
	mysqlStorage, err := storage.NewMySQLStorage("mocksvr:lujing00@tcp(localhost:3306)/mocksvr")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer mysqlStorage.Close()

	mockService := service.NewMockService(mysqlStorage)
	stubHandler := handler.NewStubHandler(mockService)
	httpHandler := handler.NewHTTPHandler(mockService)

	go startStubManagementServer(stubHandler)
	go startHTTPMockServer(httpHandler)
	go startGRPCMockServer(mockService)

	select {}
}
