package gitlab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ai-hermes/buglens-v2/internal/config"
)

type Error struct{ Message string }

func (e *Error) Error() string { return e.Message }

type Client struct {
	cfg    config.Config
	client *http.Client
}

func New(cfg config.Config) *Client {
	return &Client{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) baseURL() (string, error) {
	if strings.TrimSpace(c.cfg.GitLabURL) == "" {
		return "", &Error{Message: "Missing env var: GITLAB_URL"}
	}
	return strings.TrimRight(c.cfg.GitLabURL, "/"), nil
}

func (c *Client) token() (string, error) {
	if strings.TrimSpace(c.cfg.GitLabToken) == "" {
		return "", &Error{Message: "Missing env var: GITLAB_TOKEN"}
	}
	return c.cfg.GitLabToken, nil
}

func (c *Client) resolveProjectID(projectID string) (string, error) {
	if strings.TrimSpace(projectID) != "" {
		return projectID, nil
	}
	if strings.TrimSpace(c.cfg.GitLabProjectID) != "" {
		return c.cfg.GitLabProjectID, nil
	}
	return "", &Error{Message: "Missing project_id argument and GITLAB_PROJECT_ID env var"}
}

func sanitizePagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 1
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func (c *Client) projectAPIBase(projectID string) (string, error) {
	id, err := c.resolveProjectID(projectID)
	if err != nil {
		return "", err
	}
	base, err := c.baseURL()
	if err != nil {
		return "", err
	}
	return base + "/api/v4/projects/" + url.PathEscape(id), nil
}

func (c *Client) globalAPIBase() (string, error) {
	base, err := c.baseURL()
	if err != nil {
		return "", err
	}
	return base + "/api/v4", nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, params map[string]string, body any) (*http.Response, error) {
	token, err := c.token()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		if strings.TrimSpace(v) == "" {
			continue
		}
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(b))
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
		return nil, mapStatusErr(resp.StatusCode, string(b))
	}
	return resp, nil
}

func mapStatusErr(status int, body string) error {
	switch status {
	case 401:
		return &Error{Message: fmt.Sprintf("GitLab auth failed (401): %s", body)}
	case 403:
		return &Error{Message: fmt.Sprintf("GitLab permission denied (403): %s", body)}
	case 404:
		return &Error{Message: fmt.Sprintf("GitLab resource not found (404): %s", body)}
	case 429:
		return &Error{Message: fmt.Sprintf("GitLab rate limited (429): %s", body)}
	case 400:
		return &Error{Message: fmt.Sprintf("GitLab validation error (400): %s", body)}
	default:
		return &Error{Message: fmt.Sprintf("GitLab API error status=%d: %s", status, body)}
	}
}

func decodeJSONBody(resp *http.Response, out any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) ListProjects(ctx context.Context, search string, membership, owned bool, page, perPage int) (map[string]any, error) {
	base, err := c.globalAPIBase()
	if err != nil {
		return nil, err
	}
	page, perPage = sanitizePagination(page, perPage)
	resp, err := c.do(ctx, http.MethodGet, base+"/projects", map[string]string{
		"search":     search,
		"membership": strconv.FormatBool(membership),
		"owned":      strconv.FormatBool(owned),
		"simple":     "true",
		"page":       strconv.Itoa(page),
		"per_page":   strconv.Itoa(perPage),
	}, nil)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := decodeJSONBody(resp, &items); err != nil {
		return nil, err
	}
	projects := make([]map[string]any, 0, len(items))
	for _, item := range items {
		projects = append(projects, map[string]any{
			"id":                  item["id"],
			"name":                item["name"],
			"path_with_namespace": item["path_with_namespace"],
			"default_branch":      item["default_branch"],
			"visibility":          item["visibility"],
			"web_url":             item["web_url"],
			"last_activity_at":    item["last_activity_at"],
		})
	}
	return map[string]any{"projects": projects, "page": page, "per_page": perPage, "count": len(projects)}, nil
}

func (c *Client) SearchProjects(ctx context.Context, query string, page, perPage int) (map[string]any, error) {
	return c.ListProjects(ctx, query, false, false, page, perPage)
}

func (c *Client) GetProject(ctx context.Context, projectID string) (map[string]any, error) {
	base, err := c.globalAPIBase()
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/projects/"+url.PathEscape(projectID), nil, nil)
	if err != nil {
		return nil, err
	}
	var item map[string]any
	if err := decodeJSONBody(resp, &item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Client) GetFile(ctx context.Context, filePath, ref, projectID string) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/repository/files/"+url.PathEscape(filePath), map[string]string{"ref": ref}, nil)
	if err != nil {
		return nil, err
	}
	var item map[string]any
	if err := decodeJSONBody(resp, &item); err != nil {
		return nil, err
	}
	decoded := ""
	if enc, _ := item["encoding"].(string); enc == "base64" {
		content, _ := item["content"].(string)
		if b, err := base64.StdEncoding.DecodeString(content); err == nil {
			decoded = string(b)
		}
	}
	return map[string]any{
		"file_path":      pickStr(item, "file_path", filePath),
		"ref":            ref,
		"size":           item["size"],
		"encoding":       item["encoding"],
		"content":        decoded,
		"blob_id":        item["blob_id"],
		"last_commit_id": item["last_commit_id"],
	}, nil
}

func (c *Client) ListBranches(ctx context.Context, projectID, search string) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/repository/branches", map[string]string{"search": search}, nil)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := decodeJSONBody(resp, &items); err != nil {
		return nil, err
	}
	return map[string]any{"branches": items, "count": len(items)}, nil
}

func (c *Client) GetBranch(ctx context.Context, branch, projectID string) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/repository/branches/"+url.PathEscape(branch), nil, nil)
	if err != nil {
		return nil, err
	}
	var item map[string]any
	if err := decodeJSONBody(resp, &item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindPageCode(ctx context.Context, filePath string, line int, branch string, contextLines int, projectID string) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/repository/files/"+url.PathEscape(filePath)+"/raw", map[string]string{"ref": branch}, nil)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	total := len(lines)
	if line < 1 || line > max(total, 1) {
		return nil, &Error{Message: fmt.Sprintf("Line out of range: %d for file with %d lines", line, total)}
	}
	start := max(0, line-contextLines-1)
	end := min(total, line+contextLines)
	outLines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		prefix := "    "
		if i+1 == line {
			prefix = ">>> "
		}
		outLines = append(outLines, fmt.Sprintf("%s%4d: %s", prefix, i+1, lines[i]))
	}
	lang := mapExtToLang(path.Ext(filePath))
	return map[string]any{
		"file_path":   filePath,
		"branch":      branch,
		"line":        line,
		"total_lines": total,
		"context":     map[string]any{"start": start + 1, "end": end, "lines": outLines},
		"language":    lang,
	}, nil
}

func (c *Client) GetCommits(ctx context.Context, filePath, branch, since string, limit int, projectID string) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	params := map[string]string{"path": filePath, "ref_name": branch, "per_page": strconv.Itoa(min(max(limit, 1), 100))}
	if strings.TrimSpace(since) != "" {
		params["since"] = since + "T00:00:00Z"
	}
	resp, err := c.do(ctx, http.MethodGet, base+"/repository/commits", params, nil)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := decodeJSONBody(resp, &items); err != nil {
		return nil, err
	}
	commits := make([]map[string]any, 0, len(items))
	for _, it := range items {
		id, _ := it["id"].(string)
		if len(id) > 8 {
			id = id[:8]
		}
		commits = append(commits, map[string]any{
			"id":             id,
			"author":         it["author_name"],
			"author_email":   it["author_email"],
			"message":        it["title"],
			"committed_date": it["committed_date"],
			"web_url":        it["web_url"],
		})
	}
	var owner any
	if len(commits) > 0 {
		owner = commits[0]["author"]
	}
	return map[string]any{"commits": commits, "suggested_owner": owner, "file_path": filePath, "branch": branch}, nil
}

func (c *Client) doProjectJSON(ctx context.Context, method, projectID, endpoint string, params map[string]string, body any) (map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, method, base+endpoint, params, body)
	if err != nil {
		return nil, err
	}
	var item map[string]any
	if err := decodeJSONBody(resp, &item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Client) doProjectJSONArray(ctx context.Context, method, projectID, endpoint string, params map[string]string, body any) ([]map[string]any, error) {
	base, err := c.projectAPIBase(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, method, base+endpoint, params, body)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := decodeJSONBody(resp, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func pickStr(item map[string]any, key, fallback string) string {
	if v, ok := item[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func mapExtToLang(ext string) string {
	switch strings.ToLower(ext) {
	case ".tsx", ".ts":
		return "typescript"
	case ".jsx", ".js":
		return "javascript"
	case ".vue":
		return "vue"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".less":
		return "less"
	default:
		return "text"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func toProjectID(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	switch n := v.(type) {
	case float64:
		return strconv.Itoa(int(n))
	case int:
		return strconv.Itoa(n)
	default:
		return ""
	}
}

func toInt(v any, fallback int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	return fallback
}

func toBool(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
	}
	return fallback
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		out = append(out, fmt.Sprintf("%v", item))
	}
	return out
}

// Generic helpers for endpoint wrappers.
func ListWithPagination(ctx context.Context, c *Client, projectID, endpoint string, state, search string, page, perPage int, extra map[string]string) (map[string]any, error) {
	page, perPage = sanitizePagination(page, perPage)
	params := map[string]string{"page": strconv.Itoa(page), "per_page": strconv.Itoa(perPage), "state": state, "search": search}
	for k, v := range extra {
		params[k] = v
	}
	items, err := c.doProjectJSONArray(ctx, http.MethodGet, projectID, endpoint, params, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "page": page, "per_page": perPage, "count": len(items)}, nil
}
