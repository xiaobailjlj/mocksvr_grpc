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
		zap.String("response_code", req.ResponseCode),
		zap.String("response_header", req.ResponseHeader),
		zap.String("response_body", req.ResponseBody),
		zap.String("owner", req.Owner),
		zap.String("description", req.Description),
		zap.String("meta", req.Meta),
		zap.Int("rules_count", len(req.Rules)))

	for i, rule := range req.Rules {
		logger.Info("Rule details",
			zap.Int("rule_index", i),
			zap.Int32("match_type", rule.MatchType),
			zap.String("match_rule", rule.MatchRule),
			zap.String("response_code", rule.ResponseCode),
			zap.String("response_header", rule.ResponseHeader),
			zap.String("response_body", rule.ResponseBody),
			zap.Int32("delay_time", rule.DelayTime),
			zap.String("description", rule.Description),
			zap.String("meta", rule.Meta))
	}

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

func (s *MockService) GetAllMockUrls(ctx context.Context, req *pb.GetAllMockUrlsRequest) (*pb.GetAllMockUrlsResponse, error) {
	logger.Info("Getting all mock URLs",
		zap.String("owner", req.Owner),
		zap.Int32("page", req.Page),
		zap.Int32("pageSize", req.PageSize))

	interfaces, total, err := s.storage.GetAllMockUrls(ctx, req.Owner, int(req.Page), int(req.PageSize))
	if err != nil {
		logger.Error("Failed to get mock URLs",
			zap.Error(err))
		return nil, err
	}

	pbUrls := make([]*pb.MockUrl, 0, len(interfaces))
	for _, iface := range interfaces {
		// Convert interface header to JSON string
		headerJSON, err := json.Marshal(iface.ResponseHeader)
		if err != nil {
			logger.Error("Failed to marshal interface response header",
				zap.Int64("interface_id", iface.ID),
				zap.Error(err))
			return nil, err
		}

		// Get rules for each interface
		rules, err := s.storage.GetRulesByInterfaceID(ctx, iface.ID)
		if err != nil {
			logger.Error("Failed to get rules for interface",
				zap.Int64("interface_id", iface.ID),
				zap.Error(err))
			return nil, err
		}

		// Convert rules to protobuf format
		pbRules := make([]*pb.Rule, 0, len(rules))
		for _, rule := range rules {
			// Convert rule header to JSON string
			ruleHeaderJSON, err := json.Marshal(rule.ResponseHeader)
			if err != nil {
				logger.Error("Failed to marshal rule response header",
					zap.Int64("interface_id", iface.ID),
					zap.Error(err))
				return nil, err
			}

			pbRules = append(pbRules, &pb.Rule{
				MatchType:      rule.MatchType,
				MatchRule:      rule.MatchRule,
				ResponseCode:   rule.ResponseCode,
				ResponseHeader: string(ruleHeaderJSON),
				ResponseBody:   rule.ResponseBody,
				DelayTime:      rule.DelayTime,
				Description:    rule.Description,
				Meta:           rule.Meta,
			})
		}

		pbUrls = append(pbUrls, &pb.MockUrl{
			Id:             iface.ID,
			Url:            iface.URL,
			ResponseCode:   iface.ResponseCode,
			ResponseHeader: string(headerJSON),
			ResponseBody:   iface.ResponseBody,
			Owner:          iface.Owner,
			Description:    iface.Description,
			Meta:           iface.Meta,
			Rules:          pbRules,
		})
	}

	logger.Info("Retrieved mock URLs successfully",
		zap.Int("count", len(pbUrls)),
		zap.Int("total", total))

	return &pb.GetAllMockUrlsResponse{
		Success:     true,
		Message:     "Mock URLs retrieved successfully",
		Urls:        pbUrls,
		Total:       int32(total),
		CurrentPage: req.Page,
		PageSize:    req.PageSize,
	}, nil
}
