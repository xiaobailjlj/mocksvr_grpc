package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
	"google.golang.org/grpc"
)

func startGRPCServer(mockService *service.MockService) {
	lis, err := net.Listen("tcp", ":7002")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterMockServerServer(s, mockService)
	log.Printf("gRPC server listening on :7002")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func startHTTPProxy(mockService *service.MockService) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body string
		if r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			body = string(bodyBytes)
		}

		resp, err := mockService.GetMockResponse(context.Background(), &pb.MockRequest{
			Url:         r.URL.Path,
			RequestBody: body,
			QueryParams: r.URL.RawQuery,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.ResponseHeader != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(resp.ResponseHeader), &headers); err == nil {
				for k, v := range headers {
					w.Header().Set(k, v)
				}
			}
		}

		if code, err := strconv.Atoi(resp.ResponseCode); err == nil {
			w.WriteHeader(code)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Write([]byte(resp.ResponseBody))
	})

	log.Printf("HTTP server listening on :7001")
	if err := http.ListenAndServe(":7001", nil); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

func main() {
	mysqlStorage, err := storage.NewMySQLStorage("user:password@tcp(localhost:3306)/mockdb")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer mysqlStorage.Close()

	mockService := service.NewMockService(mysqlStorage)

	go startGRPCServer(mockService)
	startHTTPProxy(mockService)
}
