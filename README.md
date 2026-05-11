# lgtv-mcp

A [Model Context Protocol](https://modelcontextprotocol.io/) server that
lets Claude (and any other MCP client) control your LG webOS TV.

The server itself is a stateless stdio process — it speaks MCP on stdin/stdout
and translates tool calls into HTTP requests against a local
[`lgtv-core`](https://github.com/deepanshutr/lgtv-core) daemon.

```
Claude ──MCP/stdio──▶ lgtv-mcp ──HTTP──▶ lgtv-core ──TLS/WS──▶ LG TV
```

## Install

```bash
go install github.com/deepanshutr/lgtv-mcp/cmd/lgtv-mcp@latest
```

Or download a release binary.

## Configure Claude

Add to your Claude MCP config (e.g. `~/.config/claude/mcp.json` or via
`claude mcp add`):

```jsonc
{
  "mcpServers": {
    "lgtv": {
      "command": "lgtv-mcp",
      "env": {
        "LGTV_CORE_URL": "http://127.0.0.1:8765"
      }
    }
  }
}
```

Restart Claude. You should see `lgtv` listed in `/mcp`.

## Tools exposed

| Tool | Args | What it does |
|------|------|--------------|
| `tv_wake` | — | Sends WoL, waits up to 12 s for TV. |
| `tv_power_off` | — | Standby. |
| `tv_state` | — | Returns volume, current app, mute state, etc. |
| `tv_set_volume` | `level: 0-100` | Absolute volume. |
| `tv_volume_delta` | `delta: int` | Relative volume. |
| `tv_mute` | `on: bool` | |
| `tv_launch_app` | `id: string` | App ID (use `tv_state` to discover). |
| `tv_press_key` | `name: string` | `HOME`, `BACK`, `UP`, `DOWN`, `LEFT`, `RIGHT`, `ENTER`, `EXIT`, `PLAY`, `PAUSE`, etc. |

## Example session

```
> Turn the TV on and open YouTube
[Claude calls tv_wake, then tv_launch_app(id="youtube.leanback.v4")]
Done. YouTube is launching.

> What's playing?
[Claude calls tv_state]
Current app: youtube.leanback.v4, volume 14, unmuted.

> Lower the volume by 3
[Claude calls tv_volume_delta(delta=-3)]
Volume is now 11.
```

## Configuration

| Env | Default | Meaning |
|-----|---------|---------|
| `LGTV_CORE_URL` | `http://127.0.0.1:8765` | Where `lgtv-core` is listening |
| `LGTV_MCP_LOG` | unset | If set, logs MCP traffic to this file path |

## License

MIT — see `LICENSE`.
