package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// CreateStub handles the legacy HTTP request (kept for compatibility)
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

// CreateStubGin handles the Gin version of the create stub request
func (h *StubHandler) CreateStubGin(c *gin.Context) {
	var req model.StubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	headerJSON, err := json.Marshal(req.ResponseHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid header format"})
		return
	}

	pbRules := make([]*pb.Rule, 0, len(req.Rules))
	for _, rule := range req.Rules {
		ruleHeaderJSON, err := json.Marshal(rule.ResponseHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule header format"})
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

	resp, err := h.mockService.SetMockUrl(c, pbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteStub handles the legacy HTTP request (kept for compatibility)
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

// DeleteStubGin handles the Gin version of the delete stub request
func (h *StubHandler) DeleteStubGin(c *gin.Context) {
	urlIdStr := c.Query("url_id")
	urlId, err := strconv.ParseInt(urlIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url_id"})
		return
	}

	pbReq := &pb.DeleteStubRequest{
		Id: urlId,
	}

	resp, err := h.mockService.DeleteStub(c, pbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllStubs handles the legacy HTTP request (kept for compatibility)
func (h *StubHandler) GetAllStubs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	owner := r.URL.Query().Get("owner")
	keyword := r.URL.Query().Get("keyword")
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
		Keyword:  keyword,
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

// GetAllStubsGin handles the Gin version of get all stubs request
func (h *StubHandler) GetAllStubsGin(c *gin.Context) {
	// Parse query parameters
	owner := c.Query("owner")
	keyword := c.Query("keyword")

	// Default pagination values
	page := 1
	pageSize := 10

	// Parse pagination parameters if provided
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			pageSize = s
		}
	}

	pbReq := &pb.GetAllMockUrlsRequest{
		Owner:    owner,
		Keyword:  keyword,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.mockService.GetAllMockUrls(c, pbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetRules handles the legacy HTTP request (kept for compatibility)
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

// GetRulesGin handles the Gin version of get rules request
func (h *StubHandler) GetRulesGin(c *gin.Context) {
	urlIdStr := c.Query("url_id")
	urlId, err := strconv.ParseInt(urlIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url_id"})
		return
	}

	pbReq := &pb.GetRuleRequest{
		Id: urlId,
	}

	resp, err := h.mockService.GetRule(c, pbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
