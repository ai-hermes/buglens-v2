package arms

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/monitoring"

	gmcp "github.com/mark3labs/mcp-go/mcp"
)

type MCPToolRegistrar func(
	name string,
	description string,
	handler func(context.Context, map[string]any) (map[string]any, error),
	schemaOpts ...gmcp.ToolOption,
)

func RegisterMCPTools(cfg config.Config, register MCPToolRegistrar) {
	sls := monitoring.NewSLSClient(cfg)
	armsClient := New(cfg)
	service := monitoring.NewAtomicService(sls, armsClient)

	rumListAppsOpts := []gmcp.ToolOption{
		gmcp.WithString("page_token"),
		gmcp.WithNumber("page_size"),
	}
	register("arms_rum_list_apps", "List ARMS RUM apps with normalized fields.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return service.ArmsListRUMApps(
				ctx,
				strArg(args, "page_token", ""),
				intArg(args, "page_size", 100),
			), nil
		},
		rumListAppsOpts...,
	)

	rumSearchErrorsOpts := []gmcp.ToolOption{
		gmcp.WithString("project"),
		gmcp.WithString("logstore"),
		gmcp.WithString("last"),
		gmcp.WithNumber("time_from_ms"),
		gmcp.WithNumber("time_to_ms"),
		gmcp.WithString("query"),
		gmcp.WithString("event_type"),
		gmcp.WithString("app_id"),
		gmcp.WithArray("app_types", gmcp.Items(map[string]any{"type": "string"})),
		gmcp.WithString("exception_message"),
		gmcp.WithString("keyword"),
		gmcp.WithString("page_token"),
		gmcp.WithNumber("page_size"),
		gmcp.WithBoolean("reverse"),
	}
	register("arms_rum_search_errors", "Search ARMS frontend (RUM) errors by structured filters or raw query.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			project, logstore, err := resolveRUMTarget(cfg, args)
			if err != nil {
				return nil, err
			}

			var fromPtr, toPtr *int64
			if v, ok := int64Arg(args["time_from_ms"]); ok {
				fromPtr = &v
			}
			if v, ok := int64Arg(args["time_to_ms"]); ok {
				toPtr = &v
			}

			var lastPtr *string
			if l := strings.TrimSpace(strArg(args, "last", "")); l != "" {
				lastPtr = &l
			}

			fromMS, toMS, err := monitoring.ResolveTimeRange(fromPtr, toPtr, lastPtr)
			if err != nil {
				return nil, err
			}

			query := monitoring.BuildRUMSearchQuery(
				strArg(args, "query", ""),
				strArg(args, "event_type", "exception"),
				strArg(args, "app_id", ""),
				stringSliceArg(args, "app_types"),
				strArg(args, "exception_message", ""),
				strArg(args, "keyword", ""),
			)

			payload := service.SLSSearchLogs(
				ctx,
				project,
				logstore,
				fromMS,
				toMS,
				strArg(args, "page_token", ""),
				intArg(args, "page_size", 50),
				boolArg(args, "reverse", true),
				query,
			)
			payload["query"] = query
			return payload, nil
		},
		rumSearchErrorsOpts...,
	)

	rumGetErrorContextOpts := []gmcp.ToolOption{
		gmcp.WithString("pack_id", gmcp.Required()),
		gmcp.WithString("pack_meta", gmcp.Required()),
		gmcp.WithString("project"),
		gmcp.WithString("logstore"),
		gmcp.WithNumber("back_lines"),
		gmcp.WithNumber("forward_lines"),
	}
	register("arms_rum_get_error_context", "Get ARMS/SLS error context lines around a log pack record.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			project, logstore, err := resolveRUMTarget(cfg, args)
			if err != nil {
				return nil, err
			}
			return service.SLSGetLogContext(
				ctx,
				project,
				logstore,
				strArg(args, "pack_id", ""),
				strArg(args, "pack_meta", ""),
				intArg(args, "back_lines", 30),
				intArg(args, "forward_lines", 30),
			), nil
		},
		rumGetErrorContextOpts...,
	)

	stackHandler := func(ctx context.Context, args map[string]any) (map[string]any, error) {
		payload := service.ArmsResolveExceptionStack(
			ctx,
			strArg(args, "pid", ""),
			intArg(args, "line", 0),
			intArg(args, "column", 0),
			strArg(args, "sourcemap_type", "js"),
			strArg(args, "exception_binary_images", ""),
		)
		payload["exception_stack"] = fmt.Sprintf(
			"%d,%d,20",
			intArg(args, "line", 0),
			intArg(args, "column", 0),
		)
		return payload, nil
	}

	stackOpts := []gmcp.ToolOption{
		gmcp.WithString("pid", gmcp.Required()),
		gmcp.WithNumber("line", gmcp.Required()),
		gmcp.WithNumber("column", gmcp.Required()),
		gmcp.WithString("sourcemap_type"),
		gmcp.WithString("exception_binary_images"),
	}
	register(
		"arms_rum_resolve_exception_stack",
		"Resolve frontend exception stack with source map by pid + line + column.",
		stackHandler,
		stackOpts...,
	)
	register(
		"arms_exception_stack_tool",
		"Compatibility alias of arms_rum_resolve_exception_stack.",
		stackHandler,
		stackOpts...,
	)

	errorDetailOpts := []gmcp.ToolOption{
		gmcp.WithString("app"),
		gmcp.WithString("page"),
		gmcp.WithString("version"),
		gmcp.WithString("error_message"),
		gmcp.WithString("project"),
		gmcp.WithString("logstore"),
		gmcp.WithNumber("time_from_ms"),
		gmcp.WithNumber("time_to_ms"),
		gmcp.WithString("query"),
		gmcp.WithNumber("page_size"),
	}
	register("arms_get_error_detail", "Get detailed error logs by app/page/version/message filters.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			project, logstore, err := resolveRUMTarget(cfg, args)
			if err != nil {
				return nil, err
			}

			var fromPtr, toPtr *int64
			if v, ok := int64Arg(args["time_from_ms"]); ok {
				fromPtr = &v
			}
			if v, ok := int64Arg(args["time_to_ms"]); ok {
				toPtr = &v
			}

			fromMS, toMS, err := monitoring.ResolveTimeRange(fromPtr, toPtr, nil)
			if err != nil {
				return nil, err
			}

			parts := []string{}
			for _, key := range []string{"app", "page", "version", "error_message", "query"} {
				if value := strings.TrimSpace(strArg(args, key, "")); value != "" {
					parts = append(parts, value)
				}
			}

			query := "*"
			if len(parts) > 0 {
				query = strings.Join(parts, " and ")
			}

			payload := service.SLSSearchLogs(
				ctx,
				project,
				logstore,
				fromMS,
				toMS,
				"",
				intArg(args, "page_size", 20),
				true,
				query,
			)
			payload["query"] = query
			return payload, nil
		},
		errorDetailOpts...,
	)
}

func resolveRUMTarget(cfg config.Config, args map[string]any) (string, string, error) {
	project := strArg(args, "project", "")
	logstore := strArg(args, "logstore", "")
	if project == "" {
		project = cfg.RUMSLSProject
	}
	if logstore == "" {
		logstore = cfg.RUMSLSLogstore
	}
	if project == "" || logstore == "" {
		return "", "", fmt.Errorf(
			"RUM queries require project/logstore (set params or BUGLENS_RUM_SLS_PROJECT/BUGLENS_RUM_SLS_LOGSTORE)",
		)
	}
	return project, logstore, nil
}

func strArg(args map[string]any, key, fallback string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return fallback
	}
	s := fmt.Sprintf("%v", v)
	if s == "<nil>" {
		return fallback
	}
	return s
}

func intArg(args map[string]any, key string, fallback int) int {
	v, ok := args[key]
	if !ok {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		if iv, err := strconv.Atoi(n); err == nil {
			return iv
		}
	}
	return fallback
}

func boolArg(args map[string]any, key string, fallback bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	if v, ok := args[key].(string); ok {
		if parsed, err := strconv.ParseBool(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func stringSliceArg(args map[string]any, key string) []string {
	arr, ok := args[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		out = append(out, fmt.Sprintf("%v", v))
	}
	return out
}

func int64Arg(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case int64:
		return n, true
	case string:
		if iv, err := strconv.ParseInt(n, 10, 64); err == nil {
			return iv, true
		}
	}
	return 0, false
}
