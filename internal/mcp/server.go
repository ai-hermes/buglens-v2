package mcp

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"strings"

	gmcp "github.com/mark3labs/mcp-go/mcp"
	gserver "github.com/mark3labs/mcp-go/server"
	zlog "github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Transport string

	Host           string
	Port           int
	StreamablePath string

	AllowHosts                    []string
	AllowOrigins                  []string
	DisableDNSRebindingProtection bool
	EnableDNSRebindingProtection  bool

	LogLevel string
}

func (c *ServerConfig) NormalizeSecurity() {
	c.EnableDNSRebindingProtection = !c.DisableDNSRebindingProtection
	isLocal := c.Host == "127.0.0.1" || c.Host == "localhost" || c.Host == "::1"
	if c.Transport == "streamable-http" && !isLocal && len(c.AllowHosts) == 0 && !c.DisableDNSRebindingProtection {
		c.EnableDNSRebindingProtection = false
		zlog.Warn().
			Str("host", c.Host).
			Msg("non-local bind without --allow-host, disabling dns rebinding protection")
	}
	if c.StreamablePath == "" {
		c.StreamablePath = "/mcp"
	}
}

type Server struct {
	registry *Registry
	cfg      ServerConfig
}

func NewServer(registry *Registry, cfg ServerConfig) *Server {
	return &Server{registry: registry, cfg: cfg}
}

func (s *Server) Run(ctx context.Context) error {
	mcpServer := gserver.NewMCPServer("buglens", "0.1.0", gserver.WithToolCapabilities(true))
	for _, t := range s.registry.Tools() {
		tool := t.MCPTool
		handler := t.Handler
		mcpServer.AddTool(tool, func(ctx context.Context, req gmcp.CallToolRequest) (*gmcp.CallToolResult, error) {
			args := req.GetArguments()
			if args == nil {
				args = map[string]any{}
			}
			result, err := handler(ctx, args)
			if err != nil {
				return gmcp.NewToolResultError(err.Error()), nil
			}
			out := &gmcp.CallToolResult{
				Content:           []gmcp.Content{gmcp.NewTextContent(mustJSON(result))},
				StructuredContent: result,
				IsError:           result["error"] != nil,
			}
			return out, nil
		})
	}

	switch s.cfg.Transport {
	case "stdio":
		componentLogger := zlog.With().Str("component", "mcp-go").Logger()
		return gserver.ServeStdio(
			mcpServer,
			gserver.WithErrorLogger(stdlog.New(componentLogger, "", 0)),
		)
	case "streamable-http":
		return s.serveStreamableHTTP(ctx, mcpServer)
	default:
		return fmt.Errorf("unsupported transport: %s", s.cfg.Transport)
	}
}

func (s *Server) serveStreamableHTTP(ctx context.Context, mcpServer *gserver.MCPServer) error {
	streamable := gserver.NewStreamableHTTPServer(mcpServer)

	mux := http.NewServeMux()
	mux.Handle(s.cfg.StreamablePath, s.withSecurity(streamable))

	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	zlog.Info().
		Str("host", s.cfg.Host).
		Int("port", s.cfg.Port).
		Str("path", s.cfg.StreamablePath).
		Msg("streamable endpoint ready")
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) withSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.EnableDNSRebindingProtection {
			if len(s.cfg.AllowHosts) > 0 {
				ok := false
				for _, p := range s.cfg.AllowHosts {
					if wildcardMatch(p, r.Host) {
						ok = true
						break
					}
				}
				if !ok {
					http.Error(w, "invalid host header", http.StatusBadRequest)
					return
				}
			}
			if len(s.cfg.AllowOrigins) > 0 {
				origin := r.Header.Get("Origin")
				if origin != "" {
					ok := false
					for _, p := range s.cfg.AllowOrigins {
						if wildcardMatch(p, origin) {
							ok = true
							break
						}
					}
					if !ok {
						http.Error(w, "invalid origin header", http.StatusBadRequest)
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func wildcardMatch(pattern, value string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		idx := 0
		for i, p := range parts {
			if p == "" {
				continue
			}
			next := strings.Index(value[idx:], p)
			if next < 0 {
				return false
			}
			if i == 0 && !strings.HasPrefix(value, p) && !strings.HasPrefix(pattern, "*") {
				return false
			}
			idx += next + len(p)
		}
		if !strings.HasSuffix(pattern, "*") && !strings.HasSuffix(value, parts[len(parts)-1]) {
			return false
		}
		return true
	}
	return strings.EqualFold(pattern, value)
}
