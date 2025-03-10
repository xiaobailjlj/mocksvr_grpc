package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	mockService *service.MockService
}

func NewHTTPHandler(mockService *service.MockService) *HTTPHandler {
	return &HTTPHandler{mockService: mockService}
}

// ServeMock handles the legacy HTTP request (kept for compatibility)
func (h *HTTPHandler) ServeMock(w http.ResponseWriter, r *http.Request) {
	var body string
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		body = string(bodyBytes)
	}

	resp, err := h.mockService.GetMockResponse(context.Background(), &pb.MockRequest{
		Url:         r.URL.Path,
		RequestBody: body,
		QueryParams: r.URL.RawQuery,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(resp.ResponseHeader), &headers); err == nil {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}

	logger.Info("Response with code:",
		zap.String("ResponseCode", resp.ResponseCode))

	if code, err := strconv.Atoi(resp.ResponseCode); err == nil {
		w.WriteHeader(code)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Write([]byte(resp.ResponseBody))
}

// ServeMockGin handles the Gin version of mock serving
func (h *HTTPHandler) ServeMockGin(c *gin.Context) {
	var body string
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		body = string(bodyBytes)
	}

	resp, err := h.mockService.GetMockResponse(c, &pb.MockRequest{
		Url:         c.Request.URL.Path,
		RequestBody: body,
		QueryParams: c.Request.URL.RawQuery,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(resp.ResponseHeader), &headers); err == nil {
		for k, v := range headers {
			c.Header(k, v)
		}
	}

	logger.Info("Response with code:",
		zap.String("ResponseCode", resp.ResponseCode))

	code := http.StatusOK
	if codeInt, err := strconv.Atoi(resp.ResponseCode); err == nil {
		code = codeInt
	}

	c.Data(code, "application/json", []byte(resp.ResponseBody))
}
