package handler

import (
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
	"google.golang.org/grpc"
)

func RegisterGRPCServer(s *grpc.Server, mockService *service.MockService) {
	pb.RegisterMockServerServer(s, mockService)
}
