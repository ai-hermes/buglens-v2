# buglens-v2 (Go, mcp-go)

Go rewrite of buglens MCP runtime and CLI.

## Architecture (Decoupled)

- `internal/gitlab/mcp_tools.go`: GitLab MCP tool definitions and registration
- `internal/arms/mcp_tools.go`: ARMS/RUM MCP tool definitions and registration
- `internal/mcp/server.go`: MCP exposure layer using `github.com/mark3labs/mcp-go`
- `internal/mcp/tool_registry_core.go`: tool registry and domain-tool assembly

This means adding/updating tools no longer depends on MCP transport implementation.

## Dotenv

- CLI auto-loads `.env` from the current working directory.
- Existing environment variables are not overridden by `.env`.
- Custom dotenv path can be set via `BUGLENS_DOTENV_PATH`.

## Logging

- Full logging is based on `zerolog`.
- Global default level can be set with `BUGLENS_LOG_LEVEL` (default `info`).
- `buglens mcp serve --log-level ...` can override runtime level per process.

## Commands

- `buglens mcp serve`
- `buglens mcp call --tool <name> --args-json '{}'`
- `buglens version`

## MCP transport

`buglens mcp serve` supports:

- `--transport stdio|streamable-http`
- `--host`
- `--port`
- `--streamable-path`
- `--allow-host`
- `--allow-origin`
- `--disable-dns-rebinding-protection`
- `--log-level`

## Tool coverage

- GitLab tools: 32
- ARMS/RUM tools: 6
- Total: 38

## Test

```bash
make test
# or
GOTOOLCHAIN=auto CGO_ENABLED=0 go test ./...
```

## Build

```bash
make build
# or
GOTOOLCHAIN=auto goreleaser build --snapshot --clean
```
