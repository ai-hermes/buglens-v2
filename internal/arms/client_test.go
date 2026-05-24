package arms

import (
	"context"
	"testing"

	arms20190808 "github.com/alibabacloud-go/arms-20190808/v11/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/monitoring"
)

type stubARMSAPI struct {
	appsFn      func(*arms20190808.GetRumAppsRequest, *util.RuntimeOptions) (*arms20190808.GetRumAppsResponse, error)
	searchFn    func(*arms20190808.GetRumDataForPageRequest, *util.RuntimeOptions) (*arms20190808.GetRumDataForPageResponse, error)
	stackFn     func(*arms20190808.GetRumExceptionStackRequest, *util.RuntimeOptions) (*arms20190808.GetRumExceptionStackResponse, error)
	lastAppsReq *arms20190808.GetRumAppsRequest
	lastSearch  *arms20190808.GetRumDataForPageRequest
	lastStack   *arms20190808.GetRumExceptionStackRequest
}

func (s *stubARMSAPI) GetRumAppsWithOptions(req *arms20190808.GetRumAppsRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumAppsResponse, error) {
	s.lastAppsReq = req
	if s.appsFn != nil {
		return s.appsFn(req, runtime)
	}
	return &arms20190808.GetRumAppsResponse{}, nil
}

func (s *stubARMSAPI) GetRumDataForPageWithOptions(req *arms20190808.GetRumDataForPageRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumDataForPageResponse, error) {
	s.lastSearch = req
	if s.searchFn != nil {
		return s.searchFn(req, runtime)
	}
	return &arms20190808.GetRumDataForPageResponse{}, nil
}

func (s *stubARMSAPI) GetRumExceptionStackWithOptions(req *arms20190808.GetRumExceptionStackRequest, runtime *util.RuntimeOptions) (*arms20190808.GetRumExceptionStackResponse, error) {
	s.lastStack = req
	if s.stackFn != nil {
		return s.stackFn(req, runtime)
	}
	return &arms20190808.GetRumExceptionStackResponse{}, nil
}

func newLiveTestClient(stub armsAPI) *Client {
	cfg := config.Config{
		AlibabaAccessKeyID:     "ak",
		AlibabaAccessKeySecret: "sk",
		AlibabaRegionID:        "cn-hangzhou",
	}
	c := New(cfg)
	c.api = stub
	return c
}

func TestGetRUMAppsMock(t *testing.T) {
	t.Setenv("BUGLENS_MONITORING_MOCK", "true")
	c := New(config.Config{})
	res, err := c.GetRUMApps(context.Background(), "", 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	data := res.Data.(map[string]any)
	if data["count"].(int) != 1 {
		t.Fatalf("unexpected count: %#v", data["count"])
	}
}

func TestGetRUMAppsLivePagination(t *testing.T) {
	stub := &stubARMSAPI{}
	stub.appsFn = func(_ *arms20190808.GetRumAppsRequest, _ *util.RuntimeOptions) (*arms20190808.GetRumAppsResponse, error) {
		return &arms20190808.GetRumAppsResponse{Body: &arms20190808.GetRumAppsResponseBody{
			RequestId: tea.String("req-apps"),
			AppList: []*arms20190808.GetRumAppsResponseBodyAppList{
				{Pid: tea.String("p1"), AppType: tea.String("web"), RegionId: tea.String("cn-hangzhou")},
				{Pid: tea.String("p2"), AppType: tea.String("miniapp"), RegionId: tea.String("cn-hangzhou")},
				{Pid: tea.String("p3"), AppType: tea.String("ios"), RegionId: tea.String("cn-hangzhou")},
			},
		}}, nil
	}
	c := newLiveTestClient(stub)

	res, err := c.GetRUMApps(context.Background(), "2", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stub.lastAppsReq == nil || tea.StringValue(stub.lastAppsReq.RegionId) != "cn-hangzhou" {
		t.Fatalf("region not mapped")
	}
	if res.RequestID != "req-apps" {
		t.Fatalf("unexpected request id: %s", res.RequestID)
	}
	if res.NextPageToken != "3" {
		t.Fatalf("unexpected next page token: %s", res.NextPageToken)
	}
	data := res.Data.(map[string]any)
	items := data["items"].([]map[string]any)
	if len(items) != 1 || items[0]["pid"].(string) != "p2" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestSearchRUMErrorsLiveMapping(t *testing.T) {
	stub := &stubARMSAPI{}
	stub.searchFn = func(_ *arms20190808.GetRumDataForPageRequest, _ *util.RuntimeOptions) (*arms20190808.GetRumDataForPageResponse, error) {
		return &arms20190808.GetRumDataForPageResponse{Body: &arms20190808.GetRumDataForPageResponseBody{
			RequestId: tea.String("req-search"),
			Data: &arms20190808.GetRumDataForPageResponseBodyData{
				Items: []map[string]interface{}{{"k": "v"}},
				Page:  tea.String("2"),
				Total: tea.String("5"),
			},
		}}, nil
	}
	c := newLiveTestClient(stub)

	res, err := c.SearchRUMErrors(context.Background(), 1779428866000, 1779429946000, "2", 2, "event_type: exception")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stub.lastSearch == nil {
		t.Fatalf("missing request")
	}
	if tea.Int32Value(stub.lastSearch.StartTime) != 1779428866 || tea.Int32Value(stub.lastSearch.EndTime) != 1779429946 {
		t.Fatalf("time conversion mismatch")
	}
	if tea.Int32Value(stub.lastSearch.CurrentPage) != 2 || tea.Int32Value(stub.lastSearch.PageSize) != 2 {
		t.Fatalf("pagination mismatch")
	}
	if tea.StringValue(stub.lastSearch.Query) != "event_type: exception" {
		t.Fatalf("query mismatch")
	}
	if res.NextPageToken != "3" {
		t.Fatalf("unexpected next page token: %s", res.NextPageToken)
	}
	if res.RequestID != "req-search" {
		t.Fatalf("unexpected request id: %s", res.RequestID)
	}
}

func TestGetRUMExceptionStackLiveMapping(t *testing.T) {
	stub := &stubARMSAPI{}
	stub.stackFn = func(_ *arms20190808.GetRumExceptionStackRequest, _ *util.RuntimeOptions) (*arms20190808.GetRumExceptionStackResponse, error) {
		return &arms20190808.GetRumExceptionStackResponse{Body: &arms20190808.GetRumExceptionStackResponseBody{
			RequestId: tea.String("req-stack"),
			Data: &arms20190808.GetRumExceptionStackResponseBodyData{
				Lines: []*string{tea.String("line1")},
			},
		}}, nil
	}
	c := newLiveTestClient(stub)

	res, err := c.GetRUMExceptionStack(context.Background(), "pid-1", 10, 20, "js", "{\"platform\":\"h5\"}")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stub.lastStack == nil {
		t.Fatalf("missing request")
	}
	if tea.StringValue(stub.lastStack.ExceptionStack) != "10,20,20" {
		t.Fatalf("stack mismatch")
	}
	if tea.StringValue(stub.lastStack.Pid) != "pid-1" || tea.StringValue(stub.lastStack.RegionId) != "cn-hangzhou" {
		t.Fatalf("pid/region mismatch")
	}
	if res.RequestID != "req-stack" {
		t.Fatalf("unexpected request id: %s", res.RequestID)
	}
}

func TestMapUpstreamError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code monitoring.UnifiedErrorCode
	}{
		{name: "auth", err: &tea.SDKError{Code: tea.String("InvalidAccessKeyId.NotFound"), Message: tea.String("bad key"), StatusCode: tea.Int(403)}, code: monitoring.AuthFailed},
		{name: "rate", err: &tea.SDKError{Code: tea.String("Throttling.User"), Message: tea.String("too many"), StatusCode: tea.Int(429)}, code: monitoring.RateLimited},
		{name: "invalid", err: &tea.SDKError{Code: tea.String("InvalidParameter"), Message: tea.String("invalid"), StatusCode: tea.Int(400)}, code: monitoring.InvalidParam},
		{name: "timeout", err: &tea.SDKError{Code: tea.String("InternalError"), Message: tea.String("request timeout")}, code: monitoring.Timeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ae := mapUpstreamError(tc.err)
			if ae.Code != tc.code {
				t.Fatalf("expected %s, got %s", tc.code, ae.Code)
			}
		})
	}
}
