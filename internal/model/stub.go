// internal/model/stub.go
package model

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)

type StubRequest struct {
	URL            string            `json:"url" binding:"required"`
	ResponseCode   string            `json:"response_code" binding:"required"`
	ResponseHeader map[string]string `json:"response_header" binding:"required"`
	ResponseBody   string            `json:"response_body" binding:"required"`
	Owner          string            `json:"owner" binding:"required"`
	Description    string            `json:"description"`
	Meta           string            `json:"meta"`
	Rules          []Rule            `json:"rules"`
}

type Rule struct {
	MatchType      int32             `json:"match_type" binding:"required"`
	MatchRule      string            `json:"match_rule" binding:"required"`
	ResponseCode   string            `json:"response_code" binding:"required"`
	ResponseHeader map[string]string `json:"response_header" binding:"required"`
	ResponseBody   string            `json:"response_body" binding:"required"`
	DelayTime      int32             `json:"delay_time"`
	Description    string            `json:"description"`
	Meta           string            `json:"meta"`
}

type Interface struct {
	ID             int64
	URL            string
	ResponseCode   string
	ResponseHeader map[string]string
	ResponseBody   string
	Owner          string
	Description    string
	Meta           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MockResponse struct {
	InterfaceID    int64             `json:"interface_id"`
	ResponseCode   string            `json:"response_code"`
	ResponseHeader map[string]string `json:"response_header"`
	ResponseBody   string            `json:"response_body"`
}
