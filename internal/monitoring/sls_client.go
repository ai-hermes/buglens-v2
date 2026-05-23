package monitoring

import (
	"context"
	"os"
	"strings"

	"github.com/ai-hermes/buglens-v2/internal/config"
)

type SLSClient struct {
	cfg config.Config
}

func NewSLSClient(cfg config.Config) *SLSClient { return &SLSClient{cfg: cfg} }

func (c *SLSClient) SearchLogs(ctx context.Context, project, logstore string, fromMS, toMS int64, pageToken string, pageSize int, reverse bool, query string) (*AdapterResult, error) {
	_ = ctx
	_ = pageToken
	_ = pageSize
	_ = reverse
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &AdapterResult{RequestID: "mock-sls-search", Data: map[string]any{"items": []map[string]any{{"msg": "mock", "project": project, "logstore": logstore, "query": query, "from_ms": fromMS, "to_ms": toMS}}, "count": 1}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &AdapterError{Code: AuthFailed, Message: "Missing monitoring credentials"}
	}
	return nil, &AdapterError{Code: UpstreamError, Message: "SLS live API not configured in this build; set BUGLENS_MONITORING_MOCK=true for local tests"}
}

func (c *SLSClient) GetLogContext(ctx context.Context, project, logstore, packID, packMeta string, backLines, forwardLines int) (*AdapterResult, error) {
	_ = ctx
	_ = backLines
	_ = forwardLines
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &AdapterResult{RequestID: "mock-sls-context", Data: map[string]any{"items": []map[string]any{{"pack_id": packID, "pack_meta": packMeta, "project": project, "logstore": logstore}}}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &AdapterError{Code: AuthFailed, Message: "Missing monitoring credentials"}
	}
	return nil, &AdapterError{Code: UpstreamError, Message: "SLS live API not configured in this build; set BUGLENS_MONITORING_MOCK=true for local tests"}
}
