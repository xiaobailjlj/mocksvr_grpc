// internal/service/mock_service.go
package service

import (
	"context"
	"encoding/json"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/model"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/storage"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
	"go.uber.org/zap"
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
	logger.Info("Setting mock URL",
		zap.String("url", req.Url),
		zap.String("owner", req.Owner),
		zap.Int("rules_count", len(req.Rules)))

	// Parse header JSON
	var respHeader map[string]string
	if req.ResponseHeader != "" {
		if err := json.Unmarshal([]byte(req.ResponseHeader), &respHeader); err != nil {
			logger.Error("Failed to parse response header",
				zap.String("header", req.ResponseHeader),
				zap.Error(err))
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
		logger.Error("Failed to save mock URL",
			zap.String("url", req.Url),
			zap.Error(err))
		return nil, err
	}

	logger.Info("Saved mock URL successfully",
		zap.String("url", req.Url),
		zap.Int64("interface_id", interfaceID))

	// Save rules
	for i, pbRule := range req.Rules {
		var ruleHeader map[string]string
		if pbRule.ResponseHeader != "" {
			if err := json.Unmarshal([]byte(pbRule.ResponseHeader), &ruleHeader); err != nil {
				logger.Error("Failed to parse rule response header",
					zap.Int("rule_index", i),
					zap.String("header", pbRule.ResponseHeader),
					zap.Error(err))
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
			logger.Error("Failed to save rule",
				zap.Int("rule_index", i),
				zap.Int64("interface_id", interfaceID),
				zap.Error(err))
			return nil, err
		}

		logger.Debug("Saved rule successfully",
			zap.Int("rule_index", i),
			zap.Int64("interface_id", interfaceID),
			zap.Int32("match_type", rule.MatchType))
	}

	return &pb.SetMockUrlResponse{
		Success: true,
		Message: "Mock URL created successfully",
	}, nil
}

func (s *MockService) GetMockResponse(ctx context.Context, req *pb.MockRequest) (*pb.MockResponse, error) {
	logger.Info("Getting mock response",
		zap.String("url", req.Url),
		zap.String("query_params", req.QueryParams))

	mockResp, err := s.storage.GetMockResponse(ctx, req.Url)
	if err != nil {
		logger.Error("Failed to get mock response",
			zap.String("url", req.Url),
			zap.Error(err))
		return nil, err
	}

	rules, err := s.storage.GetRules(ctx, mockResp.InterfaceID)
	if err != nil {
		logger.Error("Failed to get rules",
			zap.Int64("interface_id", mockResp.InterfaceID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("Found rules for URL",
		zap.String("url", req.Url),
		zap.Int("rules_count", len(rules)))

	for i, rule := range rules {
		matches := false
		if rule.MatchType == 1 && req.QueryParams == rule.MatchRule {
			matches = true
			logger.Debug("Query parameter rule matched",
				zap.Int("rule_index", i),
				zap.String("query_params", req.QueryParams))
		} else if rule.MatchType == 2 && req.RequestBody == rule.MatchRule {
			matches = true
			logger.Debug("Request body rule matched",
				zap.Int("rule_index", i),
				zap.String("request_body", req.RequestBody))
		}

		if matches {
			if rule.DelayTime > 0 {
				logger.Debug("Applying delay",
					zap.Int32("delay_ms", rule.DelayTime))
				time.Sleep(time.Duration(rule.DelayTime) * time.Millisecond)
			}

			headerJSON, err := json.Marshal(rule.ResponseHeader)
			if err != nil {
				logger.Error("Failed to marshal rule response header",
					zap.Int("rule_index", i),
					zap.Error(err))
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
	logger.Info("No rules matched, using default response",
		zap.String("url", req.Url))

	headerJSON, err := json.Marshal(mockResp.ResponseHeader)
	if err != nil {
		logger.Error("Failed to marshal default response header",
			zap.Error(err))
		return nil, err
	}

	return &pb.MockResponse{
		ResponseCode:   mockResp.ResponseCode,
		ResponseHeader: string(headerJSON),
		ResponseBody:   mockResp.ResponseBody,
	}, nil
}
