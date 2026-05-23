package mcp

import "testing"

func TestNormalizeSecurityNonLocalDisablesDNSProtectionWithoutAllowHost(t *testing.T) {
	cfg := ServerConfig{Transport: "streamable-http", Host: "0.0.0.0", Port: 18002, StreamablePath: "/mcp"}
	cfg.NormalizeSecurity()
	if cfg.EnableDNSRebindingProtection {
		t.Fatalf("expected dns rebinding protection disabled by default for non-local without allow-host")
	}
}

func TestNormalizeSecurityKeepsDefaultSTDIO(t *testing.T) {
	cfg := ServerConfig{Transport: "stdio", Host: "127.0.0.1", Port: 8000, StreamablePath: "/mcp"}
	cfg.NormalizeSecurity()
	if cfg.Host != "127.0.0.1" || cfg.Port != 8000 || cfg.StreamablePath != "/mcp" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
