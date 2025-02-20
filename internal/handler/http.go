package handler

import (
	"context"
	"encoding/json"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"

	"github.com/xiaobailjlj/mocksvr_grpc/internal/service"
	pb "github.com/xiaobailjlj/mocksvr_grpc/proto/mockserver"
)

type HTTPHandler struct {
	mockService *service.MockService
}

func NewHTTPHandler(mockService *service.MockService) *HTTPHandler {
	return &HTTPHandler{mockService: mockService}
}

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
