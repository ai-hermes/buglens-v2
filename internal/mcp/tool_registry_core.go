package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/ai-hermes/buglens-v2/internal/arms"
	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/gitlab"

	gmcp "github.com/mark3labs/mcp-go/mcp"
)

type ToolHandler func(context.Context, map[string]any) (map[string]any, error)

type Tool struct {
	Name        string
	Description string
	MCPTool     gmcp.Tool
	Handler     ToolHandler
}

type Registry struct {
	cfg   config.Config
	tools map[string]Tool
}

func NewRegistry(cfg config.Config) *Registry {
	r := &Registry{cfg: cfg, tools: map[string]Tool{}}

	gitlab.RegisterMCPTools(
		cfg,
		func(
			name string,
			description string,
			handler func(context.Context, map[string]any) (map[string]any, error),
			schemaOpts ...gmcp.ToolOption,
		) {
			r.registerTool(name, description, handler, schemaOpts...)
		},
	)

	arms.RegisterMCPTools(
		cfg,
		func(
			name string,
			description string,
			handler func(context.Context, map[string]any) (map[string]any, error),
			schemaOpts ...gmcp.ToolOption,
		) {
			r.registerTool(name, description, handler, schemaOpts...)
		},
	)
	return r
}

func (r *Registry) Register(t Tool) {
	if t.MCPTool.Name == "" {
		t.MCPTool = gmcp.NewTool(t.Name, gmcp.WithDescription(t.Description))
	}
	r.tools[t.Name] = t
}

func (r *Registry) registerTool(
	name string,
	description string,
	handler ToolHandler,
	schemaOpts ...gmcp.ToolOption,
) {
	opts := append([]gmcp.ToolOption{gmcp.WithDescription(description)}, schemaOpts...)
	r.Register(Tool{
		Name:        name,
		Description: description,
		MCPTool:     gmcp.NewTool(name, opts...),
		Handler:     handler,
	})
}

func (r *Registry) Tools() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) ToolNames() []string {
	tools := r.Tools()
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name)
	}
	return names
}

func (r *Registry) Call(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	result, err := tool.Handler(ctx, args)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	return result, nil
}
