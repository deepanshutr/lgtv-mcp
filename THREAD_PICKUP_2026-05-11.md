# Thread pickup — 2026-05-11 (lgtv-mcp)

## What shipped this session

Initial scaffolding of the Go MCP server (mark3labs/mcp-go, stdio
transport). Single commit: `bab9e94` Initial scaffolding.

Exposes **8 tools** to MCP clients (Claude Code, etc.):

| Tool | Maps to lgtv-core endpoint |
|---|---|
| `tv_wake` | `POST /wake` |
| `tv_power_off` | `POST /power-off` |
| `tv_state` | `GET /state` |
| `tv_set_volume` | `POST /volume` (absolute) |
| `tv_volume_delta` | `POST /volume` (relative) |
| `tv_mute` | `POST /mute` |
| `tv_launch_app` | `POST /launch` |
| `tv_press_key` | `POST /key` |

## Live state

- Binary built and committed (`lgtv-mcp`, 9.5 MB) — also at
  `~/.local/bin/lgtv-mcp`.
- **Registered with Claude Code at user scope** as `lgtv`. Verify with
  `claude mcp list`.
- Talks to `lgtv-core` over `http://127.0.0.1:8765` (configurable via
  `LGTV_CORE_URL`).

## Verification done this session

- MCP stdio smoke test (raw JSON-RPC `initialize` + `tools/list`) returns
  all 8 tools. See `feedback_mcp_stdio_smoke_test.md` in memory.
- Through Claude Code the `mcp__lgtv__tv_state` tool is reachable
  (visible in tool list at this session's start).

## Gotchas to NOT re-discover

1. Local Go env: `~/.profile` exports `GOROOT=/home/deepanshutr/go/go1.18`
   which mismatches the brew go at `/home/linuxbrew/.linuxbrew/bin/go`
   (1.26.2). Always `unset GOROOT; export
   GOPROXY=https://proxy.golang.org,direct` before any `go` command.
2. Push email must be the noreply:
   `52166434+deepanshutr@users.noreply.github.com`. Per-repo only.
3. **Don't commit the compiled binary** going forward — `bab9e94` did,
   but future tags should use goreleaser releases instead. Add `lgtv-mcp`
   to `.gitignore` if iterating locally.

## Exact resume incantation

```bash
cd ~/github.com/deepanshutr/lgtv-mcp
unset GOROOT; export GOPROXY=https://proxy.golang.org,direct
go build -o ~/.local/bin/lgtv-mcp .

# Smoke test stdio:
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | ~/.local/bin/lgtv-mcp 2>/dev/null | jq -c '.result.tools[]?.name // .'

# Confirm Claude Code registration:
claude mcp list | grep lgtv
```

## Repo state at thread-close

- Branch: `main`, up to date with `origin/main`, clean tree.
- Single commit since init.
- CI green; gitleaks green; goreleaser config in place.

## Memory references

- `project_lgtv_stack.md` — full project / sibling-repo overview
- `feedback_mcp_stdio_smoke_test.md` — how to drive any MCP server
  without a real client
