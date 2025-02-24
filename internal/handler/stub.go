package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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

func (h *StubHandler) DeleteStub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlIdStr := r.URL.Query().Get("url_id")
	urlId, err := strconv.ParseInt(urlIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid url_id", http.StatusBadRequest)
		return
	}

	pbReq := &pb.DeleteStubRequest{
		Id: urlId,
	}

	resp, err := h.mockService.DeleteStub(context.Background(), pbReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *StubHandler) GetAllStubs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	owner := r.URL.Query().Get("owner")
	page := 1
	pageSize := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			pageSize = s
		}
	}

	pbReq := &pb.GetAllMockUrlsRequest{
		Owner:    owner,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.mockService.GetAllMockUrls(context.Background(), pbReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *StubHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlIdStr := r.URL.Query().Get("url_id")
	urlId, err := strconv.ParseInt(urlIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid url_id", http.StatusBadRequest)
		return
	}

	pbReq := &pb.GetRuleRequest{
		Id: urlId,
	}

	resp, err := h.mockService.GetRule(context.Background(), pbReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
