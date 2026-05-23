package monitoring

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type UnifiedErrorCode string

const (
	AuthFailed       UnifiedErrorCode = "AUTH_FAILED"
	PermissionDenied UnifiedErrorCode = "PERMISSION_DENIED"
	RateLimited      UnifiedErrorCode = "RATE_LIMITED"
	Timeout          UnifiedErrorCode = "TIMEOUT"
	InvalidParam     UnifiedErrorCode = "INVALID_PARAM"
	UpstreamError    UnifiedErrorCode = "UPSTREAM_ERROR"
)

type UnifiedError struct {
	Code           UnifiedErrorCode `json:"code"`
	Message        string           `json:"message"`
	RequestID      string           `json:"request_id,omitempty"`
	Retriable      bool             `json:"retriable"`
	UpstreamStatus int              `json:"upstream_status,omitempty"`
	UpstreamCode   string           `json:"upstream_code,omitempty"`
	Details        map[string]any   `json:"details,omitempty"`
}

type Envelope struct {
	Success        bool          `json:"success"`
	RequestID      string        `json:"request_id,omitempty"`
	LatencyMS      int64         `json:"latency_ms"`
	Data           any           `json:"data,omitempty"`
	Error          *UnifiedError `json:"error,omitempty"`
	NextPageToken  string        `json:"next_page_token,omitempty"`
	PartialSuccess bool          `json:"partial_success,omitempty"`
}

type AdapterResult struct {
	Data           any
	RequestID      string
	NextPageToken  string
	PartialSuccess bool
}

type AdapterError struct {
	Code           UnifiedErrorCode
	Message        string
	RequestID      string
	Retriable      bool
	UpstreamStatus int
	UpstreamCode   string
	Details        map[string]any
}

func (e *AdapterError) Error() string { return e.Message }

func (e *AdapterError) ToUnifiedError() *UnifiedError {
	return &UnifiedError{
		Code:           e.Code,
		Message:        e.Message,
		RequestID:      e.RequestID,
		Retriable:      e.Retriable,
		UpstreamStatus: e.UpstreamStatus,
		UpstreamCode:   e.UpstreamCode,
		Details:        e.Details,
	}
}

func EncodePageToken(payload map[string]any) string {
	b, _ := json.Marshal(payload)
	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodePageToken(token string) (map[string]any, error) {
	if token == "" {
		return map[string]any{}, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, &AdapterError{Code: InvalidParam, Message: "Invalid page_token", Details: map[string]any{"page_token": token}}
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, &AdapterError{Code: InvalidParam, Message: "Invalid page_token", Details: map[string]any{"page_token": token}}
	}
	return out, nil
}

func Wrap(fn func() (*AdapterResult, error)) map[string]any {
	start := nowMS()
	result, err := fn()
	latency := nowMS() - start
	if err != nil {
		if ae, ok := err.(*AdapterError); ok {
			return map[string]any{"success": false, "latency_ms": latency, "request_id": ae.RequestID, "error": ae.ToUnifiedError()}
		}
		return map[string]any{"success": false, "latency_ms": latency, "error": map[string]any{"code": UpstreamError, "message": err.Error(), "retriable": false}}
	}
	return map[string]any{
		"success":         true,
		"request_id":      result.RequestID,
		"latency_ms":      latency,
		"data":            result.Data,
		"next_page_token": result.NextPageToken,
		"partial_success": result.PartialSuccess,
	}
}

func nowMS() int64 { return time.Now().UnixMilli() }
