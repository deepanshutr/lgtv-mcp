# Thread pickup — 2026-05-12

## What shipped this thread

Initial scaffold of `lgtv-mcp`: a stdio Model Context Protocol server that
lets Claude (and any MCP client) control the LG TV via the local
`lgtv-core` daemon.

- `cmd/lgtv-mcp` — entrypoint
- `internal/tools` — 8 MCP tools registered (`tv_wake`, `tv_power_off`,
  `tv_state`, `tv_set_volume`, `tv_volume_delta`, `tv_mute`, `tv_launch_app`,
  `tv_press_key`)
- `internal/core` — HTTP client for `lgtv-core` (intentional copy of
  lgtv-cli's client to keep this binary self-contained)
- Smoke test script `scripts/smoke.sh` drives the stdio handshake
- CI runs `go vet` + `go test` + the stdio smoke test; gitleaks; release
  pipeline via goreleaser; dependabot
- MIT licensed, public

## Live state on operator host

- Binary at `~/.local/bin/lgtv-mcp`
- Registered with Claude Code at **user scope**:
  ```
  claude mcp list  # → lgtv: /home/deepanshutr/.local/bin/lgtv-mcp  - ✓ Connected
  ```
- New Claude Code sessions will have `tv_*` tools available; this session
  doesn't (MCP tool list is captured at session start).

## Verified end-to-end

- `initialize` + `tools/list` over stdio → 8 tools enumerated
- `tools/call name=tv_state` against live daemon → returned full TV state
  (25 apps, 6 inputs, volume, current channel)

## Resume incantation

```bash
cd ~/github.com/deepanshutr/lgtv-mcp
go vet ./... && go test ./... -count=1
go build -o ~/.local/bin/lgtv-mcp ./cmd/lgtv-mcp
./scripts/smoke.sh ~/.local/bin/lgtv-mcp   # initialize + tools/list

# Verify MCP registration
claude mcp list | grep lgtv

# Drive a tool manually (e.g. tv_state) without Claude
{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"manual","version":"0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"tv_state","arguments":{}}}'
} | ~/.local/bin/lgtv-mcp
```

## Gotchas (don't re-discover)

- MCP stdio servers MUST NOT use stdout for logging — corrupts the JSON-RPC
  stream. Use stderr or `LGTV_MCP_LOG` to a file.
- `mark3labs/mcp-go` is the SDK in use (not Anthropic's `go-sdk`).

## Related repos

- `lgtv-core` — Python daemon
- `lgtv-cli` — sibling Go CLI/bot
