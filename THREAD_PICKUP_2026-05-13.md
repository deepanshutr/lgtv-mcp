# Thread Pickup — 2026-05-13

State of the repo when this session ended.

## What changed

Two commits this session:

1. **`6e2ba3e` Feat: tv_launch_app exposes content_id and content_target**
   — adds optional `content_id` (app-specific deep-link key like a
   YouTube video ID) and `content_target` (deep-link URL, wrapped into
   `{contentTarget: ...}` before being sent as `params.contentTarget`).
2. **`38cb102` Feat: tv_youtube_pair, tv_youtube_play, tv_youtube_status
   MCP tools** — three new tools wrapping the YouTube Lounge endpoints
   in lgtv-core (`/youtube/pair`, `/youtube/play`, `/youtube/status`).
   This is the **reliable** way to deep-link YouTube content on CX
   firmware; see the lgtv-core thread-pickup for protocol details.

Total tool count is now 11. End-to-end stdio smoke test verified:
initialize → tools/list shows all 11; tools/call on
`tv_youtube_status` round-trips through the daemon and returns
`paired=true` with the cached screen_id.

## What you need to know

### The binary cache trap

This MCP server is spawned as a child stdio process by Claude. Its tool
schema is read ONCE at session-start. The new binary is at
`~/.local/bin/lgtv-mcp` and is the latest build — but any currently-open
Claude session was started with the *old* binary and won't see the new
args.

**To pick up the new tool schema, the user must restart their Claude
session.** New sessions automatically read the upgraded binary.

This is why the binary install uses `install -m 0755` (handles
text-file-busy) rather than `cp` (fails when child is running).

### YouTube Lounge tools — DONE (`38cb102`)

All three tools are now wired and the upgraded binary is installed at
`~/.local/bin/lgtv-mcp`. Schemas:

- `tv_youtube_play(video_id: string, start_time_s?: number)` — plays
  the given video. Returns `Playing YouTube video <id>.` on success or
  surfaces daemon diagnostic on failure.
- `tv_youtube_pair(pairing_code: string)` — exchanges a 12-digit TV
  code for a permanent paired loungeToken. Dashes/spaces in the code
  are stripped daemon-side.
- `tv_youtube_status()` — returns paired/unpaired + cached IDs.

User already has a paired session as of this commit's verification
(screenId `5b563a65…`). Re-pairing happens automatically when needed
— `tv_youtube_play` will return an error like "not paired" or
"lounge token rejected" that should be surfaced to the user, who then
calls `tv_youtube_pair` with a fresh code.

### Adding new core HTTP endpoints to the Go client

Pattern (see `internal/core/client.go`):

```go
type YoutubePlayOpts struct { VideoId string; StartTimeS int }
func (c *Client) YoutubePlay(ctx context.Context, opts YoutubePlayOpts) error {
    return c.do(ctx, "POST", "/youtube/play",
        map[string]any{"video_id": opts.VideoId, "start_time_s": opts.StartTimeS},
        nil)
}
```

And in `internal/tools/tools.go`, use `mcp.WithString` + `req.RequireString`
to wire the MCP tool.

## Verified end-to-end this session

- `tv_launch_app` with `content_id` and `content_target` proven via
  daemon-direct curl (the MCP path needs a session restart to verify
  end-to-end through Claude).
- All other existing tools confirmed working after the daemon-side
  upgrades to `/app/launch` (post_keys, content_id/params).

## Things to remember

- This repo intentionally duplicates `internal/core/` with `lgtv-cli`
  (see CLAUDE.md). When extending the daemon client, update both repos
  if both should expose the same surface.
- The Go build wants `GOROOT` unset on this machine — the user has a
  stale `~/go/go1.18` directory. Use `env -u GOROOT go build` or set
  `GOROOT=""`.
