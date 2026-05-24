package arms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	arms20190808 "github.com/alibabacloud-go/arms-20190808/v11/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"

	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/monitoring"
)

type armsAPI interface {
	GetRumAppsWithOptions(request *arms20190808.GetRumAppsRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumAppsResponse, error)
	GetRumDataForPageWithOptions(request *arms20190808.GetRumDataForPageRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumDataForPageResponse, error)
	GetRumExceptionStackWithOptions(request *arms20190808.GetRumExceptionStackRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumExceptionStackResponse, error)
}

type Client struct {
	cfg    config.Config
	mu     sync.Mutex
	api    armsAPI
	newAPI func() (armsAPI, error)
}

func New(cfg config.Config) *Client {
	c := &Client{cfg: cfg}
	c.newAPI = c.defaultNewAPI
	return c
}

func (c *Client) defaultNewAPI() (armsAPI, error) {
	cfg := credential.Config{}
	if strings.TrimSpace(c.cfg.AlibabaSecurityToken) != "" {
		cfg.SetType("sts")
		cfg.SetSecurityToken(c.cfg.AlibabaSecurityToken)
	} else {
		cfg.SetType("access_key")
	}
	cfg.SetAccessKeyId(c.cfg.AlibabaAccessKeyID)
	cfg.SetAccessKeySecret(c.cfg.AlibabaAccessKeySecret)

	cred, err := credential.NewCredential(&cfg)
	if err != nil {
		return nil, err
	}

	opencfg := &openapi.Config{Credential: cred}
	opencfg.Endpoint = tea.String(resolveARMSEndpoint(c.cfg))
	return arms20190808.NewClient(opencfg)
}

func resolveARMSEndpoint(cfg config.Config) string {
	endpoint := strings.TrimSpace(cfg.ARMSEndpoint)
	region := strings.TrimSpace(cfg.AlibabaRegionID)
	if endpoint != "" && endpoint != "arms.aliyuncs.com" {
		return endpoint
	}
	if region != "" {
		return fmt.Sprintf("arms.%s.aliyuncs.com", region)
	}
	if endpoint != "" {
		return endpoint
	}
	return "arms.aliyuncs.com"
}

func (c *Client) getAPI() (armsAPI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api != nil {
		return c.api, nil
	}
	api, err := c.newAPI()
	if err != nil {
		return nil, err
	}
	c.api = api
	return c.api, nil
}

func (c *Client) regionID() string {
	region := strings.TrimSpace(c.cfg.AlibabaRegionID)
	if region == "" {
		return "cn-hangzhou"
	}
	return region
}

func (c *Client) GetRUMApps(ctx context.Context, pageToken string, pageSize int) (*monitoring.AdapterResult, error) {
	_ = ctx
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &monitoring.AdapterResult{RequestID: "mock-arms-1", Data: map[string]any{"items": []map[string]any{{"app_type": "web", "pid": "app-1", "region_id": c.regionID()}}, "count": 1, "total": 1, "page": 1, "page_size": 1}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &monitoring.AdapterError{Code: monitoring.AuthFailed, Message: "Missing monitoring credentials"}
	}

	page, err := parsePageToken(pageToken)
	if err != nil {
		return nil, err
	}
	pageSize = normalizePageSize(pageSize, 100)

	api, err := c.getAPI()
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	resp, err := api.GetRumAppsWithOptions(&arms20190808.GetRumAppsRequest{RegionId: tea.String(c.regionID())}, &util.RuntimeOptions{})
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	apps := make([]map[string]any, 0)
	requestID := ""
	if resp != nil && resp.Body != nil {
		requestID = tea.StringValue(resp.Body.RequestId)
		for _, app := range resp.Body.AppList {
			if app == nil {
				continue
			}
			apps = append(apps, map[string]any{
				"app_type":          tea.StringValue(app.AppType),
				"create_time":       app.CreateTime,
				"description":       tea.StringValue(app.Description),
				"endpoint":          tea.StringValue(app.Endpoint),
				"is_subscription":   tea.BoolValue(app.IsSubscription),
				"name":              tea.StringValue(app.Name),
				"pid":               tea.StringValue(app.Pid),
				"region_id":         tea.StringValue(app.RegionId),
				"resource_group_id": tea.StringValue(app.ResourceGroupId),
				"sls_project":       tea.StringValue(app.SlsProject),
				"sls_logstore":      tea.StringValue(app.SlsLogstore),
				"status":            tea.StringValue(app.Status),
				"type":              tea.StringValue(app.Type),
			})
		}
	}

	total := len(apps)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := apps[start:end]

	nextPageToken := ""
	if end < total {
		nextPageToken = strconv.Itoa(page + 1)
	}
	return &monitoring.AdapterResult{
		RequestID: requestID,
		Data: map[string]any{
			"items":     items,
			"count":     len(items),
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
		NextPageToken: nextPageToken,
	}, nil
}

func (c *Client) SearchRUMErrors(ctx context.Context, fromMS, toMS int64, pageToken string, pageSize int, query string) (*monitoring.AdapterResult, error) {
	_ = ctx
	if strings.EqualFold(os.Getenv("BUGLENS_MONITORING_MOCK"), "true") {
		return &monitoring.AdapterResult{RequestID: "mock-rum-search-1", Data: map[string]any{"items": []map[string]any{{"msg": "mock-rum-error", "query": query, "from_ms": fromMS, "to_ms": toMS}}, "count": 1, "total": 1, "page": 1, "page_size": 1}}, nil
	}
	if !c.cfg.HasMonitoringCredentials() {
		return nil, &monitoring.AdapterError{Code: monitoring.AuthFailed, Message: "Missing monitoring credentials"}
	}
	if fromMS <= 0 || toMS <= 0 || fromMS > toMS {
		return nil, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "invalid time range", Details: map[string]any{"time_from_ms": fromMS, "time_to_ms": toMS}}
	}

	page, err := parsePageToken(pageToken)
	if err != nil {
		return nil, err
	}
	pageSize = normalizePageSize(pageSize, 50)
	if strings.TrimSpace(query) == "" {
		query = "* and event_type: exception"
	}

	api, err := c.getAPI()
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	startSec, err := msToSecInt32(fromMS)
	if err != nil {
		return nil, err
	}
	endSec, err := msToSecInt32(toMS)
	if err != nil {
		return nil, err
	}

	req := &arms20190808.GetRumDataForPageRequest{
		RegionId:    tea.String(c.regionID()),
		StartTime:   tea.Int32(startSec),
		EndTime:     tea.Int32(endSec),
		CurrentPage: tea.Int32(int32(page)),
		PageSize:    tea.Int32(int32(pageSize)),
		Query:       tea.String(query),
	}
	resp, err := api.GetRumDataForPageWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	requestID := ""
	items := make([]map[string]any, 0)
	total := 0
	responsePage := page
	completion := ""
	if resp != nil && resp.Body != nil {
		requestID = tea.StringValue(resp.Body.RequestId)
		if resp.Body.Data != nil {
			for _, item := range resp.Body.Data.Items {
				converted := map[string]any{}
				for k, v := range item {
					converted[k] = v
				}
				items = append(items, converted)
			}
			if parsed, perr := strconv.Atoi(strings.TrimSpace(tea.StringValue(resp.Body.Data.Total))); perr == nil {
				total = parsed
			}
			if parsed, perr := strconv.Atoi(strings.TrimSpace(tea.StringValue(resp.Body.Data.Page))); perr == nil && parsed > 0 {
				responsePage = parsed
			}
			completion = strings.ToLower(strings.TrimSpace(tea.StringValue(resp.Body.Data.Completion)))
		}
	}
	if total == 0 && len(items) > 0 {
		total = len(items)
	}

	nextPageToken := ""
	if completion != "true" && (total == 0 || responsePage*pageSize < total) {
		nextPageToken = strconv.Itoa(responsePage + 1)
	}

	return &monitoring.AdapterResult{
		RequestID: requestID,
		Data: map[string]any{
			"items":     items,
			"count":     len(items),
			"total":     total,
			"page":      responsePage,
			"page_size": pageSize,
		},
		NextPageToken: nextPageToken,
	}, nil
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

	api, err := c.getAPI()
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	resp, err := api.GetRumExceptionStackWithOptions(&arms20190808.GetRumExceptionStackRequest{
		Pid:                   tea.String(pid),
		ExceptionStack:        tea.String(stack),
		ExceptionBinaryImages: tea.String(exceptionBinaryImages),
		RegionId:              tea.String(c.regionID()),
		SourcemapType:         tea.String(sourcemapType),
	}, &util.RuntimeOptions{})
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	requestID := ""
	result := map[string]any{}
	if resp != nil && resp.Body != nil {
		requestID = tea.StringValue(resp.Body.RequestId)
		if dataMap, derr := modelToMap(resp.Body.Data); derr == nil {
			result = dataMap
		}
	}

	return &monitoring.AdapterResult{
		RequestID: requestID,
		Data: map[string]any{
			"result":          result,
			"exception_stack": stack,
		},
	}, nil
}

func parsePageToken(token string) (int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 1, nil
	}
	if page, err := strconv.Atoi(token); err == nil {
		if page <= 0 {
			return 0, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "page_token must be >= 1", Details: map[string]any{"page_token": token}}
		}
		return page, nil
	}

	decoded, err := monitoring.DecodePageToken(token)
	if err != nil {
		return 0, err
	}
	for _, key := range []string{"page", "current_page", "currentPage"} {
		if v, ok := decoded[key]; ok {
			page, ok := toInt(v)
			if !ok || page <= 0 {
				return 0, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "invalid page_token", Details: map[string]any{"page_token": token}}
			}
			return page, nil
		}
	}
	return 0, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "invalid page_token", Details: map[string]any{"page_token": token}}
}

func normalizePageSize(v int, fallback int) int {
	if v <= 0 {
		return fallback
	}
	if v > 1000 {
		return 1000
	}
	return v
}

func msToSecInt32(ms int64) (int32, error) {
	sec := ms / 1000
	if sec <= 0 {
		return 0, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "timestamp must be positive", Details: map[string]any{"timestamp_ms": ms}}
	}
	if sec > int64(^uint32(0)>>1) {
		return 0, &monitoring.AdapterError{Code: monitoring.InvalidParam, Message: "timestamp out of range", Details: map[string]any{"timestamp_ms": ms}}
	}
	return int32(sec), nil
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return i, true
		}
	}
	return 0, false
}

func modelToMap(v any) (map[string]any, error) {
	if v == nil {
		return map[string]any{}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func mapUpstreamError(err error) *monitoring.AdapterError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*monitoring.AdapterError); ok {
		return ae
	}

	message := err.Error()
	code := monitoring.UpstreamError
	status := 0
	upstreamCode := ""
	retriable := false
	requestID := ""
	details := map[string]any{}

	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		upstreamCode = strings.TrimSpace(tea.StringValue(sdkErr.Code))
		status = tea.IntValue(sdkErr.StatusCode)
		if strings.TrimSpace(tea.StringValue(sdkErr.Message)) != "" {
			message = tea.StringValue(sdkErr.Message)
		}
		if raw := strings.TrimSpace(tea.StringValue(sdkErr.Data)); raw != "" {
			details["sdk_data"] = raw
			if parsed, perr := parseSDKData(raw); perr == nil {
				details["sdk_data_json"] = parsed
				if rid := extractRequestID(parsed); rid != "" {
					requestID = rid
				}
			}
		}
	}

	lowerCode := strings.ToLower(upstreamCode)
	lowerMsg := strings.ToLower(message)
	switch {
	case status == 401 || status == 403 || strings.Contains(lowerCode, "accesskey") || strings.Contains(lowerCode, "signature") || strings.Contains(lowerCode, "forbidden") || strings.Contains(lowerCode, "permission"):
		code = monitoring.AuthFailed
	case status == 429 || strings.Contains(lowerCode, "thrott") || strings.Contains(lowerCode, "toomany"):
		code = monitoring.RateLimited
		retriable = true
	case status == 400 || strings.Contains(lowerCode, "invalid") || strings.Contains(lowerCode, "missing") || strings.Contains(lowerCode, "parameter"):
		code = monitoring.InvalidParam
	case strings.Contains(lowerMsg, "timeout") || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		code = monitoring.Timeout
		retriable = true
	}

	if len(details) == 0 {
		details = nil
	}
	return &monitoring.AdapterError{
		Code:           code,
		Message:        message,
		RequestID:      requestID,
		Retriable:      retriable,
		UpstreamStatus: status,
		UpstreamCode:   upstreamCode,
		Details:        details,
	}
}

func parseSDKData(raw string) (map[string]any, error) {
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func extractRequestID(data map[string]any) string {
	for _, key := range []string{"RequestId", "requestId", "request_id"} {
		if v, ok := data[key]; ok {
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" {
				return s
			}
		}
	}
	return ""
}
