package gitlab

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func ListMergeRequests(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	extra := map[string]string{
		"source_branch": strArg(args, "source_branch", ""),
		"target_branch": strArg(args, "target_branch", ""),
	}
	return ListWithPagination(ctx, c, toProjectID(args["project_id"]), "/merge_requests", strArg(args, "state", "opened"), strArg(args, "search", ""), intArg(args, "page", 1), intArg(args, "per_page", 20), extra)
}

func GetMergeRequest(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	return c.doProjectJSON(ctx, http.MethodGet, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid), nil, nil)
}

func CreateMergeRequest(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	payload := map[string]any{
		"title":                strArg(args, "title", ""),
		"source_branch":        strArg(args, "source_branch", ""),
		"target_branch":        strArg(args, "target_branch", ""),
		"description":          strArg(args, "description", ""),
		"draft":                boolArg(args, "draft", false),
		"remove_source_branch": boolArg(args, "remove_source_branch", false),
	}
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/merge_requests", nil, payload)
}

func UpdateMergeRequest(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	payload := map[string]any{}
	for _, k := range []string{"title", "description", "target_branch", "state_event"} {
		if v := strArg(args, k, ""); strings.TrimSpace(v) != "" {
			payload[k] = v
		}
	}
	return c.doProjectJSON(ctx, http.MethodPut, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid), nil, payload)
}

func MergeMergeRequest(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	payload := map[string]any{
		"merge_when_pipeline_succeeds": boolArg(args, "merge_when_pipeline_succeeds", false),
		"should_remove_source_branch":  boolArg(args, "should_remove_source_branch", false),
		"squash":                       boolArg(args, "squash", false),
	}
	return c.doProjectJSON(ctx, http.MethodPut, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid)+"/merge", nil, payload)
}

func GetMRChanges(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	return c.doProjectJSON(ctx, http.MethodGet, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid)+"/changes", nil, nil)
}

func GetMRDiscussions(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid)+"/discussions", nil, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"discussions": items, "count": len(items)}, nil
}

func CreateMRNote(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	payload := map[string]any{"body": strArg(args, "body", "")}
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/merge_requests/"+strconv.Itoa(iid)+"/notes", nil, payload)
}

func ListIssues(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	extra := map[string]string{
		"labels":            strArg(args, "labels", ""),
		"assignee_username": strArg(args, "assignee_username", ""),
	}
	return ListWithPagination(ctx, c, toProjectID(args["project_id"]), "/issues", strArg(args, "state", "opened"), strArg(args, "search", ""), intArg(args, "page", 1), intArg(args, "per_page", 20), extra)
}

func GetIssue(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	return c.doProjectJSON(ctx, http.MethodGet, toProjectID(args["project_id"]), "/issues/"+strconv.Itoa(iid), nil, nil)
}

func CreateIssue(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	payload := map[string]any{
		"title":       strArg(args, "title", ""),
		"description": strArg(args, "description", ""),
	}
	if labels := stringSliceArg(args, "labels"); len(labels) > 0 {
		payload["labels"] = strings.Join(labels, ",")
	}
	if assignee := strArg(args, "assignee", ""); assignee != "" {
		payload["assignee_ids"] = []string{assignee}
	}
	if ms := strArg(args, "milestone", ""); ms != "" {
		payload["milestone_id"] = ms
	}
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/issues", nil, payload)
}

func UpdateIssue(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	payload := map[string]any{}
	for _, k := range []string{"title", "description", "state_event"} {
		if v := strArg(args, k, ""); strings.TrimSpace(v) != "" {
			payload[k] = v
		}
	}
	if labels := stringSliceArg(args, "labels"); len(labels) > 0 {
		payload["labels"] = strings.Join(labels, ",")
	}
	return c.doProjectJSON(ctx, http.MethodPut, toProjectID(args["project_id"]), "/issues/"+strconv.Itoa(iid), nil, payload)
}

func CreateIssueNote(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	payload := map[string]any{"body": strArg(args, "body", "")}
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/issues/"+strconv.Itoa(iid)+"/notes", nil, payload)
}

func ListIssueNotes(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	iid := intArg(args, "iid", 0)
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, toProjectID(args["project_id"]), "/issues/"+strconv.Itoa(iid)+"/notes", nil, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"notes": items, "count": len(items)}, nil
}

func ListPipelines(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	page, perPage := sanitizePagination(intArg(args, "page", 1), intArg(args, "per_page", 20))
	params := map[string]string{
		"page":     strconv.Itoa(page),
		"per_page": strconv.Itoa(perPage),
		"ref":      strArg(args, "ref", ""),
		"status":   strArg(args, "status", ""),
	}
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, toProjectID(args["project_id"]), "/pipelines", params, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"pipelines": items, "page": page, "per_page": perPage, "count": len(items)}, nil
}

func GetPipeline(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	id := intArg(args, "pipeline_id", 0)
	return c.doProjectJSON(ctx, http.MethodGet, toProjectID(args["project_id"]), "/pipelines/"+strconv.Itoa(id), nil, nil)
}

func RetryPipeline(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	id := intArg(args, "pipeline_id", 0)
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/pipelines/"+strconv.Itoa(id)+"/retry", nil, nil)
}

func CancelPipeline(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	id := intArg(args, "pipeline_id", 0)
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/pipelines/"+strconv.Itoa(id)+"/cancel", nil, nil)
}

func ListPipelineJobs(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	id := intArg(args, "pipeline_id", 0)
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, toProjectID(args["project_id"]), "/pipelines/"+strconv.Itoa(id)+"/jobs", nil, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"jobs": items, "count": len(items)}, nil
}

func GetJobLog(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	id := intArg(args, "job_id", 0)
	base, err := c.projectAPIBase(toProjectID(args["project_id"]))
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/jobs/"+strconv.Itoa(id)+"/trace", nil, nil)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	text := string(raw)
	preview := text
	truncated := false
	if len(preview) > 8000 {
		preview = preview[:8000]
		truncated = true
	}
	return map[string]any{"job_id": id, "trace_preview": preview, "truncated": truncated}, nil
}

func ListLabels(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	page, perPage := sanitizePagination(intArg(args, "page", 1), intArg(args, "per_page", 100))
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, toProjectID(args["project_id"]), "/labels", map[string]string{"page": strconv.Itoa(page), "per_page": strconv.Itoa(perPage)}, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"labels": items, "count": len(items), "page": page, "per_page": perPage}, nil
}

func CreateLabel(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	payload := map[string]any{"name": strArg(args, "name", ""), "color": strArg(args, "color", ""), "description": strArg(args, "description", "")}
	return c.doProjectJSON(ctx, http.MethodPost, toProjectID(args["project_id"]), "/labels", nil, payload)
}

func UpdateLabel(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	payload := map[string]any{"name": strArg(args, "name", "")}
	if v := strArg(args, "new_name", ""); v != "" {
		payload["new_name"] = v
	}
	if v := strArg(args, "color", ""); v != "" {
		payload["color"] = v
	}
	if v := strArg(args, "description", ""); v != "" {
		payload["description"] = v
	}
	return c.doProjectJSON(ctx, http.MethodPut, toProjectID(args["project_id"]), "/labels", nil, payload)
}

func DeleteLabel(ctx context.Context, c *Client, args map[string]any) (map[string]any, error) {
	base, err := c.projectAPIBase(toProjectID(args["project_id"]))
	if err != nil {
		return nil, err
	}
	name := strArg(args, "name", "")
	_, err = c.do(ctx, http.MethodDelete, base+"/labels", map[string]string{"name": name}, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"deleted": true, "name": name}, nil
}

func strArg(args map[string]any, key, fallback string) string {
	v, ok := args[key]
	if !ok {
		return fallback
	}
	s := fmt.Sprintf("%v", v)
	if strings.EqualFold(s, "<nil>") {
		return fallback
	}
	return s
}

func intArg(args map[string]any, key string, fallback int) int {
	return toInt(args[key], fallback)
}

func boolArg(args map[string]any, key string, fallback bool) bool {
	return toBool(args[key], fallback)
}

func stringSliceArg(args map[string]any, key string) []string {
	return toStringSlice(args[key])
}
