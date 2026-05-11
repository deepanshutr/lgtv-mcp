# CLAUDE.md — lgtv-mcp

Context for future Claude sessions working in this repo.

## What this is

An MCP (Model Context Protocol) server. Speaks stdio JSON-RPC to its parent
process (Claude, typically) and translates tool invocations into HTTP calls
against the local [`lgtv-core`](https://github.com/deepanshutr/lgtv-core)
daemon.

## Architecture

```
Parent (Claude) ──stdio─▶ lgtv-mcp ──HTTP─▶ lgtv-core ──WSS─▶ LG TV
```

The server is **stateless** — it owns no TV connection of its own. Each tool
invocation is a fresh HTTP call to `lgtv-core`. If `lgtv-core` is down, every
tool fails with a connection error; that's intentional — clear failure mode.

## Layout

```
cmd/lgtv-mcp/main.go    entrypoint, MCP server loop
internal/tools/         tool definitions + handlers
internal/core/          HTTP client for lgtv-core (vendored copy of the
                        same code as in lgtv-cli; intentional duplication
                        to keep this binary self-contained)
```

## Design rules

1. **Stateless tools.** Never cache TV state in the MCP process — `lgtv-core`
   has the truth, query it.
2. **Tool descriptions are user-facing prompts.** They guide the model on
   when to call each tool. Keep them short, factual, and free of
   implementation hints.
3. **Errors propagate cleanly.** `lgtv-core` returns 502 with `{"detail":...}`
   on TV-level failures — surface the detail string in the MCP error, don't
   wrap it.
4. **No TV identifiers in this repo.** All TV-specific config lives in
   `lgtv-core`. This binary only knows the daemon's URL.

## Gotchas

- MCP stdio servers MUST NOT write to stdout for anything other than MCP
  protocol frames. Use stderr for logs, or write to a file via `LGTV_MCP_LOG`.
  A single stray `fmt.Println` will corrupt the JSON-RPC stream.
- Tool argument schemas matter: Claude reads them to decide which tool to
  call and how to fill the args. A vague description like "do something
  with the TV" will lead to the wrong tool being picked.
- When developing, smoke-test with `printf '...' | lgtv-mcp` — drive the
  stdio loop directly with `initialize` then `tools/list`. (Standard MCP
  protocol; see `scripts/smoke.sh`.)

## Building locally

```bash
go build ./cmd/lgtv-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}' | ./lgtv-mcp
```

## Testing

```bash
go test ./...
```

## CI/CD

- `.github/workflows/ci.yml` — `go vet`, `staticcheck`, `go test ./...` on every PR.
- `.github/workflows/release.yml` — on tag, `goreleaser` builds binaries.
- `gitleaks` runs on every push.

## Related

- `lgtv-core` — the Python daemon this MCP server talks to.
- `lgtv-cli` — sibling Go CLI / Telegram bot; also a `lgtv-core` consumer.
