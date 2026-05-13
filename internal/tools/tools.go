// Package tools registers MCP tools against an mcp-go server.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deepanshutr/lgtv-mcp/internal/core"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register attaches all TV-control tools to the server.
func Register(s *server.MCPServer, client *core.Client) {
	s.AddTool(mcp.NewTool("tv_wake",
		mcp.WithDescription("Wake the TV from standby. Sends a Wake-on-LAN magic packet and waits for the TV to be reachable. Use this first if the TV may be off."),
	), wrap(func(ctx context.Context, _ mcp.CallToolRequest) (string, error) {
		if err := client.Wake(ctx); err != nil {
			return "", err
		}
		return "TV is awake.", nil
	}))

	s.AddTool(mcp.NewTool("tv_power_off",
		mcp.WithDescription("Put the TV into standby."),
	), wrap(func(ctx context.Context, _ mcp.CallToolRequest) (string, error) {
		if err := client.PowerOff(ctx); err != nil {
			return "", err
		}
		return "TV is now off.", nil
	}))

	s.AddTool(mcp.NewTool("tv_state",
		mcp.WithDescription("Get the current TV state: volume, mute, current app, available apps, available inputs."),
	), wrap(func(ctx context.Context, _ mcp.CallToolRequest) (string, error) {
		s, err := client.State(ctx)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(s, "", "  ")
		return string(b), nil
	}))

	s.AddTool(mcp.NewTool("tv_set_volume",
		mcp.WithDescription("Set the TV volume to an absolute level (0-100)."),
		mcp.WithNumber("level", mcp.Required(), mcp.Description("Volume level, 0-100")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		level, err := req.RequireInt("level")
		if err != nil {
			return "", err
		}
		if err := client.VolumeAbsolute(ctx, level); err != nil {
			return "", err
		}
		return fmt.Sprintf("Volume set to %d.", level), nil
	}))

	s.AddTool(mcp.NewTool("tv_volume_delta",
		mcp.WithDescription("Change the TV volume by a relative amount (positive or negative)."),
		mcp.WithNumber("delta", mcp.Required(), mcp.Description("Volume delta (e.g. -3 to lower by 3 steps)")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		delta, err := req.RequireInt("delta")
		if err != nil {
			return "", err
		}
		if err := client.VolumeDelta(ctx, delta); err != nil {
			return "", err
		}
		return fmt.Sprintf("Volume changed by %+d.", delta), nil
	}))

	s.AddTool(mcp.NewTool("tv_mute",
		mcp.WithDescription("Mute or unmute the TV."),
		mcp.WithBoolean("on", mcp.Required(), mcp.Description("true to mute, false to unmute")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		on, err := req.RequireBool("on")
		if err != nil {
			return "", err
		}
		if err := client.Mute(ctx, on); err != nil {
			return "", err
		}
		if on {
			return "Muted.", nil
		}
		return "Unmuted.", nil
	}))

	s.AddTool(mcp.NewTool("tv_launch_app",
		mcp.WithDescription("Launch an app on the TV by app ID. Use tv_state to see available app IDs. Optionally deep-link into a specific piece of content via `content_id` (e.g. a YouTube video ID like 'dQw4w9WgXcQ') or `content_target` (a deep-link URL such as 'https://www.youtube.com/tv?v=...' for the YouTube app, or any URL for the 'com.webos.app.browser' app)."),
		mcp.WithString("id", mcp.Required(), mcp.Description("App ID, e.g. 'netflix' or 'youtube.leanback.v4'")),
		mcp.WithString("content_id", mcp.Description("Optional deep-link content key for the target app (YouTube: video ID; Netflix: title ID)")),
		mcp.WithString("content_target", mcp.Description("Optional deep-link URL. Sent as params.contentTarget — used by YouTube and the LG browser")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return "", err
		}
		opts := &core.LaunchAppOpts{
			ContentID: req.GetString("content_id", ""),
		}
		if target := req.GetString("content_target", ""); target != "" {
			opts.Params = map[string]any{"contentTarget": target}
		}
		if opts.ContentID == "" && opts.Params == nil {
			opts = nil
		}
		if err := client.LaunchApp(ctx, id, opts); err != nil {
			return "", err
		}
		return "Launched " + id + ".", nil
	}))

	s.AddTool(mcp.NewTool("tv_press_key",
		mcp.WithDescription("Press a remote-control key. Valid names: HOME, BACK, MENU, UP, DOWN, LEFT, RIGHT, ENTER, EXIT, PLAY, PAUSE, STOP, RED, GREEN, YELLOW, BLUE."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Key name (e.g. HOME, BACK, UP)")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return "", err
		}
		if err := client.PressKey(ctx, name); err != nil {
			return "", err
		}
		return "Pressed " + name + ".", nil
	}))

	s.AddTool(mcp.NewTool("tv_youtube_play",
		mcp.WithDescription("Play a specific YouTube video on the TV using the paired Lounge protocol. Reliable for deep-linking on LG WebOS (especially CX-era firmware where the standard launch surface ignores video IDs). Requires one-time pairing via tv_youtube_pair first; if not paired, returns an error with instructions. Fully autonomous — no profile-picker dance, no key presses needed."),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("11-character YouTube video ID (the 'v=' query param from a YouTube URL, e.g. 'jNQXAC9IVRw' for 'Me at the zoo')")),
		mcp.WithNumber("start_time_s", mcp.Description("Optional start offset in seconds (default 0)")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		videoID, err := req.RequireString("video_id")
		if err != nil {
			return "", err
		}
		startTime := req.GetInt("start_time_s", 0)
		if err := client.YoutubePlay(ctx, videoID, startTime); err != nil {
			return "", err
		}
		return "Playing YouTube video " + videoID + ".", nil
	}))

	s.AddTool(mcp.NewTool("tv_youtube_pair",
		mcp.WithDescription("One-time pair the daemon with the TV's YouTube app via a 12-digit TV code. To get the code: on the TV, open YouTube, press up to reveal the left nav, go to Settings (gear), select 'Link with TV code'. Pass the displayed 12-digit code (dashes/spaces optional) to this tool. The resulting paired loungeToken is stored locally and reused for all future tv_youtube_play calls."),
		mcp.WithString("pairing_code", mcp.Required(), mcp.Description("The 12-digit code displayed on the TV (dashes and spaces are stripped, so '893-949-732-447' or '893949732447' both work)")),
	), wrap(func(ctx context.Context, req mcp.CallToolRequest) (string, error) {
		code, err := req.RequireString("pairing_code")
		if err != nil {
			return "", err
		}
		out, err := client.YoutubePair(ctx, code)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return "Paired:\n" + string(b), nil
	}))

	s.AddTool(mcp.NewTool("tv_youtube_status",
		mcp.WithDescription("Check whether the daemon has cached YouTube Lounge pairing credentials. Returns whether paired plus the cached screen/device IDs. Use this to decide whether tv_youtube_pair needs to run first."),
	), wrap(func(ctx context.Context, _ mcp.CallToolRequest) (string, error) {
		out, err := client.YoutubeStatus(ctx)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return string(b), nil
	}))
}

// wrap turns a plain (text, error) handler into the mcp-go signature.
func wrap(fn func(context.Context, mcp.CallToolRequest) (string, error)) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := fn(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	}
}
