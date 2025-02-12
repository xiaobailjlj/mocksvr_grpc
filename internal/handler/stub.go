package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/xiaobailjlj/mocksvr_grpc/internal/model"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
)

type StubHandler struct {
	mockService *service.MockService
}

func NewStubHandler(mockService *service.MockService) *StubHandler {
	return &StubHandler{mockService: mockService}
}

func (h *StubHandler) CreateStub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.StubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	headerJSON, err := json.Marshal(req.ResponseHeader)
	if err != nil {
		http.Error(w, "Invalid header format", http.StatusBadRequest)
		return
	}

	pbRules := make([]*pb.Rule, 0, len(req.Rules))
	for _, rule := range req.Rules {
		ruleHeaderJSON, err := json.Marshal(rule.ResponseHeader)
		if err != nil {
			http.Error(w, "Invalid rule header format", http.StatusBadRequest)
			return
		}

		pbRules = append(pbRules, &pb.Rule{
			MatchType:      rule.MatchType,
			MatchRule:      rule.MatchRule,
			ResponseCode:   rule.ResponseCode,
			ResponseHeader: string(ruleHeaderJSON),
			ResponseBody:   rule.ResponseBody,
			DelayTime:      rule.DelayTime,
			Description:    rule.Description,
		})
	}

	pbReq := &pb.SetMockUrlRequest{
		Url:            req.URL,
		ResponseCode:   req.ResponseCode,
		ResponseHeader: string(headerJSON),
		ResponseBody:   req.ResponseBody,
		Owner:          req.Owner,
		Description:    req.Description,
		Rules:          pbRules,
	}

	resp, err := h.mockService.SetMockUrl(context.Background(), pbReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
