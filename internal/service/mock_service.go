// internal/service/mock_service.go
package service

import (
	"context"
	"encoding/json"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/model"
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
	// Parse header JSON
	var respHeader map[string]string
	if req.ResponseHeader != "" {
		if err := json.Unmarshal([]byte(req.ResponseHeader), &respHeader); err != nil {
			return nil, err
		}
	}

	// Save main interface
	interfaceID, err := s.storage.SaveMockUrl(
		ctx,
		req.Url,
		req.ResponseCode,
		respHeader,
		req.ResponseBody,
		req.Owner,
		req.Description,
		req.Meta,
	)
	if err != nil {
		return nil, err
	}

	// Save rules
	for _, pbRule := range req.Rules {
		var ruleHeader map[string]string
		if pbRule.ResponseHeader != "" {
			if err := json.Unmarshal([]byte(pbRule.ResponseHeader), &ruleHeader); err != nil {
				return nil, err
			}
		}

		rule := &model.Rule{
			MatchType:      pbRule.MatchType,
			MatchRule:      pbRule.MatchRule,
			ResponseCode:   pbRule.ResponseCode,
			ResponseHeader: ruleHeader,
			ResponseBody:   pbRule.ResponseBody,
			DelayTime:      pbRule.DelayTime,
			Description:    pbRule.Description,
			Meta:           pbRule.Meta,
		}

		if err := s.storage.SaveRule(ctx, interfaceID, rule); err != nil {
			return nil, err
		}
	}

	return &pb.SetMockUrlResponse{
		Success: true,
		Message: "Mock URL created successfully",
	}, nil
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

			headerJSON, err := json.Marshal(rule.ResponseHeader)
			if err != nil {
				return nil, err
			}

			return &pb.MockResponse{
				ResponseCode:   rule.ResponseCode,
				ResponseHeader: string(headerJSON),
				ResponseBody:   rule.ResponseBody,
			}, nil
		}
	}

	// If no rules match, return default response
	headerJSON, err := json.Marshal(mockResp.ResponseHeader)
	if err != nil {
		return nil, err
	}

	return &pb.MockResponse{
		ResponseCode:   mockResp.ResponseCode,
		ResponseHeader: string(headerJSON),
		ResponseBody:   mockResp.ResponseBody,
	}, nil
}
