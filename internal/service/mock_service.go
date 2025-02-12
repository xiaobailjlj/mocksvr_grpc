package service

import (
	"context"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
	"time"
)

type MockService struct {
	pb.UnimplementedMockServerServer
	storage *storage.MySQLStorage
}

func NewMockService(storage *storage.MySQLStorage) *MockService {
	return &MockService{storage: storage}
}

func (s *MockService) SetMockUrl(ctx context.Context, req *pb.SetMockUrlRequest) (*pb.SetMockUrlResponse, error) {
	interfaceID, err := s.storage.SaveMockUrl(ctx, req.Url, req.ResponseCode, req.ResponseHeader, req.ResponseBody, req.Owner, req.Description)
	if err != nil {
		return nil, err
	}

	for _, rule := range req.Rules {
		err := s.storage.SaveRule(ctx, interfaceID, rule.MatchType, rule.MatchRule, rule.ResponseCode, rule.ResponseHeader, rule.ResponseBody, rule.DelayTime, rule.Description)
		if err != nil {
			return nil, err
		}
	}

	return &pb.SetMockUrlResponse{Success: true, Message: "Mock URL created successfully"}, nil
}

func (s *MockService) GetMockResponse(ctx context.Context, req *pb.MockRequest) (*pb.MockResponse, error) {
	mockResp, err := s.storage.GetMockResponse(ctx, req.Url)
	if err != nil {
		return nil, err
	}

	rules, err := s.storage.GetRules(ctx, mockResp.InterfaceID)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		matches := false
		if rule.MatchType == 1 && req.QueryParams == rule.MatchRule {
			matches = true
		} else if rule.MatchType == 2 && req.RequestBody == rule.MatchRule {
			matches = true
		}

		if matches {
			if rule.DelayTime > 0 {
				time.Sleep(time.Duration(rule.DelayTime) * time.Millisecond)
			}
			return &pb.MockResponse{
				ResponseCode:   rule.ResponseCode,
				ResponseHeader: rule.ResponseHeader,
				ResponseBody:   rule.ResponseBody,
			}, nil
		}
	}

	return &pb.MockResponse{
		ResponseCode:   mockResp.ResponseCode,
		ResponseHeader: mockResp.ResponseHeader,
		ResponseBody:   mockResp.ResponseBody,
	}, nil
}
