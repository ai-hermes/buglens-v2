package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ai-hermes/buglens-v2/internal/config"
	"github.com/ai-hermes/buglens-v2/internal/mcp"

	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	m := &cobra.Command{Use: "mcp", Short: "MCP tools"}
	m.AddCommand(newMCPServeCmd())
	m.AddCommand(newMCPCallCmd())
	return m
}

func newMCPServeCmd() *cobra.Command {
	cfg := mcp.ServerConfig{
		Transport:      "stdio",
		Host:           "127.0.0.1",
		Port:           8000,
		StreamablePath: "/mcp",
		LogLevel:       getEnvDefault("BUGLENS_MCP_LOG_LEVEL", "INFO"),
	}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "run mcp server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setLogLevel(cfg.LogLevel); err != nil {
				return err
			}
			cfg.NormalizeSecurity()
			registry := mcp.NewRegistry(config.FromEnv())
			server := mcp.NewServer(registry, cfg)
			zlog.Info().
				Int("tools", len(registry.Tools())).
				Str("transport", cfg.Transport).
				Str("log_level", strings.ToLower(cfg.LogLevel)).
				Msg("starting mcp server")
			return server.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&cfg.Transport, "transport", getEnvDefault("BUGLENS_MCP_TRANSPORT", "stdio"), "stdio|streamable-http")
	cmd.Flags().StringVar(&cfg.Host, "host", getEnvDefault("BUGLENS_MCP_HOST", "127.0.0.1"), "http bind host")
	cmd.Flags().IntVar(&cfg.Port, "port", getEnvDefaultInt("BUGLENS_MCP_PORT", 8000), "http bind port")
	cmd.Flags().StringVar(&cfg.StreamablePath, "streamable-path", getEnvDefault("BUGLENS_MCP_STREAMABLE_HTTP_PATH", "/mcp"), "streamable http path")
	cmd.Flags().StringSliceVar(&cfg.AllowHosts, "allow-host", nil, "allowed host header (repeatable)")
	cmd.Flags().StringSliceVar(&cfg.AllowOrigins, "allow-origin", nil, "allowed origin header (repeatable)")
	cmd.Flags().BoolVar(&cfg.DisableDNSRebindingProtection, "disable-dns-rebinding-protection", false, "disable dns rebinding protection")
	cmd.Flags().StringVar(
		&cfg.LogLevel,
		"log-level",
		getEnvDefault("BUGLENS_LOG_LEVEL", cfg.LogLevel),
		"trace|debug|info|warn|error|fatal|panic",
	)
	return cmd
}

func newMCPCallCmd() *cobra.Command {
	var toolName string
	var argsJSON string

	cmd := &cobra.Command{
		Use:   "call",
		Short: "call a tool locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(toolName) == "" {
				return errors.New("--tool is required")
			}
			parsed := map[string]any{}
			if strings.TrimSpace(argsJSON) != "" {
				if err := json.Unmarshal([]byte(argsJSON), &parsed); err != nil {
					return fmt.Errorf("invalid --args-json: %w", err)
				}
			}

			registry := mcp.NewRegistry(config.FromEnv())
			result, err := registry.Call(context.Background(), toolName, parsed)
			if err != nil {
				return err
			}
			bytes, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(bytes))
			return nil
		},
	}
	cmd.Flags().StringVar(&toolName, "tool", "", "tool name")
	cmd.Flags().StringVar(&argsJSON, "args-json", "{}", "json object arguments")
	return cmd
}
