package arms

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/monitoring"
)

type Client struct {
	cfg config.Config
}

func New(cfg config.Config) *Client { return &Client{cfg: cfg} }

func (c *Client) GetRUMApps(ctx context.Context, pageToken string, pageSize int) (*monitoring.AdapterResult, error) {
	_ = ctx
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &monitoring.AdapterResult{RequestID: "mock-arms-1", Data: map[string]any{"items": []map[string]any{{"app_type": "web", "pid": "app-1", "region_id": c.cfg.AlibabaRegionID}}, "total": 1}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &monitoring.AdapterError{Code: monitoring.AuthFailed, Message: "Missing monitoring credentials"}
	}
	return nil, &monitoring.AdapterError{Code: monitoring.UpstreamError, Message: "ARMS live API not configured in this build; set BUGLENS_MONITORING_MOCK=true for local tests", Retriable: false}
}

func (c *Client) GetRUMExceptionStack(ctx context.Context, pid string, line, column int, sourcemapType, exceptionBinaryImages string) (*monitoring.AdapterResult, error) {
	_ = ctx
	if strings.TrimSpace(pid) == "" {
		return nil, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "pid is required"}
	}
	if sourcemapType == "" {
		sourcemapType = "js"
	}
	if exceptionBinaryImages == "" {
		exceptionBinaryImages = `{"platform":"h5"}`
	}
	stack := fmt.Sprintf("%d,%d,20", line, column)
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &monitoring.AdapterResult{RequestID: "mock-stack-1", Data: map[string]any{"result": map[string]any{"resolved": true, "sourcemap_type": sourcemapType, "exception_binary_images": exceptionBinaryImages}, "exception_stack": stack}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &monitoring.AdapterError{Code: monitoring.AuthFailed, Message: "Missing monitoring credentials"}
	}
	return nil, &monitoring.AdapterError{Code: monitoring.UpstreamError, Message: "ARMS live API not configured in this build; set BUGLENS_MONITORING_MOCK=true for local tests", Retriable: false}
}
