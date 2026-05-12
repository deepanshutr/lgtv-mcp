// Package core is a small HTTP client for the lgtv-core daemon.
// Intentional duplicate of lgtv-cli/internal/core — keeps lgtv-mcp self-contained.
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("call %s %s: status %d: %s", method, path, resp.StatusCode, string(data))
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) Wake(ctx context.Context) error    { return c.do(ctx, "POST", "/wake", nil, nil) }
func (c *Client) PowerOff(ctx context.Context) error { return c.do(ctx, "POST", "/power/off", nil, nil) }

func (c *Client) State(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	return out, c.do(ctx, "GET", "/state", nil, &out)
}

func (c *Client) VolumeAbsolute(ctx context.Context, level int) error {
	return c.do(ctx, "POST", "/volume", map[string]any{"level": level}, nil)
}
func (c *Client) VolumeDelta(ctx context.Context, delta int) error {
	return c.do(ctx, "POST", "/volume", map[string]any{"delta": delta}, nil)
}
func (c *Client) Mute(ctx context.Context, on bool) error {
	return c.do(ctx, "POST", "/mute", map[string]any{"on": on}, nil)
}
// LaunchAppOpts are optional deep-link extras for /app/launch.
// At most one of ContentID or Params should be set; both are passed
// through to the daemon, which forwards to aiowebostv's
// launch_app_with_content_id / launch_app_with_params helpers.
type LaunchAppOpts struct {
	ContentID string         // app-specific deep-link key (e.g. YouTube video ID)
	Params    map[string]any // arbitrary params payload (e.g. {"contentTarget": "..."})
}

func (c *Client) LaunchApp(ctx context.Context, id string, opts *LaunchAppOpts) error {
	body := map[string]any{"id": id}
	if opts != nil {
		if opts.ContentID != "" {
			body["content_id"] = opts.ContentID
		}
		if opts.Params != nil {
			body["params"] = opts.Params
		}
	}
	return c.do(ctx, "POST", "/app/launch", body, nil)
}
func (c *Client) PressKey(ctx context.Context, name string) error {
	return c.do(ctx, "POST", "/key", map[string]any{"name": name}, nil)
}
