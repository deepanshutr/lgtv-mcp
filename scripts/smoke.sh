#!/usr/bin/env bash
# Drive lgtv-mcp through the basic MCP handshake to confirm it starts cleanly.
# Doesn't require a real TV — only checks the MCP wire works.
set -euo pipefail
BIN="${1:-./lgtv-mcp}"
{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
} | "$BIN" | head -3
