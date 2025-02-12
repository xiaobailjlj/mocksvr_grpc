package model

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)

type StubRequest struct {
	URL            string            `json:"url"`
	ResponseCode   string            `json:"response_code"`
	ResponseHeader map[string]string `json:"response_header"`
	ResponseBody   string            `json:"response_body"`
	Owner          string            `json:"owner"`
	Description    string            `json:"description"`
	Meta           string            `json:"meta"`
	Rules          []Rule            `json:"rules"`
}

type Rule struct {
	MatchType      int32             `json:"match_type"`
	MatchRule      string            `json:"match_rule"`
	ResponseCode   string            `json:"response_code"`
	ResponseHeader map[string]string `json:"response_header"`
	ResponseBody   string            `json:"response_body"`
	DelayTime      int32             `json:"delay_time"`
	Description    string            `json:"description"`
	Meta           string            `json:"meta"`
}
