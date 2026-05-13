# Thread Pickup — 2026-05-13

State of the repo when this session ended.

## What changed

One commit this session: **`6e2ba3e` Feat: tv_launch_app exposes
content_id and content_target**.

The `tv_launch_app` tool now takes two optional extras:
- `content_id` — app-specific deep-link key (YouTube video ID,
  Netflix title ID, etc.)
- `content_target` — deep-link URL, wrapped into `{contentTarget: ...}`
  before being sent to the daemon as `params.contentTarget`. Used by
  the LG browser and (some firmware versions of) YouTube.

Build verified, schema verified via the stdio smoke test:
```
echo '{"jsonrpc":"2.0",…"tools/list"}' | lgtv-mcp
```
returns `tv_launch_app` with `content_id` + `content_target` as optional
string properties.

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

### Not yet wired: YouTube Lounge tools

`lgtv-core` got a major upgrade today: full YouTube Lounge protocol
support for true deep-link playback on CX-firmware TVs that ignore
SSAP launch deep-links. See `lgtv-core/THREAD_PICKUP_2026-05-13.md`
for the protocol details.

The daemon exposes three new endpoints:
- `POST /youtube/pair {"pairing_code": "..."}`
- `POST /youtube/play {"video_id": "..."}`
- `GET /youtube/status`

These should be wrapped as MCP tools `tv_youtube_pair`, `tv_youtube_play`,
`tv_youtube_status`. **This work is NOT yet done in lgtv-mcp.** Next
session candidate. The new tools should:

- Take `video_id` (required string) for `tv_youtube_play`, optional
  `start_time_s` (int, default 0).
- Take `pairing_code` (required string) for `tv_youtube_pair`. The
  description should explain where to get the code:
  *"YouTube on TV → Settings → Link with TV code"*.
- Surface daemon errors verbatim — if pairing expired, the error from
  `/youtube/play` is the user-facing diagnostic ("not paired" or
  "lounge token rejected").

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
