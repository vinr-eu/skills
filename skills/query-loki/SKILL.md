---
name: query-loki
description: Query a Grafana Loki instance using LogQL. Use when the user wants to search, tail, or filter logs from Loki. Accepts a Loki base URL, a LogQL query string, optional time range, optional bearer token, and an optional result limit.
metadata:
  author: vinr
  version: "1.0"
---

## Overview

This skill runs a prebuilt Go binary (`query-loki`) that calls the Loki HTTP API and prints matching log lines to
stdout. The binary is downloaded from GitHub Releases on first use and cached locally.

## Parameters

| Parameter   | Flag      | Env var      | Required | Default    | Notes                             |
|-------------|-----------|--------------|----------|------------|-----------------------------------|
| Loki URL    | `--url`   | `LOKI_URL`   | Yes      | —          | Base URL, e.g. `http://loki:3100` |
| LogQL query | `--query` | —            | Yes      | —          | e.g. `{job="nginx"} \|= "error"`  |
| Start time  | `--from`  | —            | No       | 1 hour ago | RFC3339 or Unix nanoseconds       |
| End time    | `--to`    | —            | No       | now        | RFC3339 or Unix nanoseconds       |
| Limit       | `--limit` | —            | No       | 100        | Max log lines returned            |
| Auth token  | `--token` | `LOKI_TOKEN` | No       | —          | Bearer token                      |
| Raw JSON    | `--json`  | —            | No       | false      | Print full Loki JSON response     |

Flags take precedence over env vars.

## Step 0 – Check for required config

Before doing anything else, check whether the environment variables are already set:

```bash
echo "LOKI_URL=${LOKI_URL}" && echo "LOKI_TOKEN=${LOKI_TOKEN}"
```

- If `LOKI_URL` is set → proceed directly to Step 1.
- If `LOKI_URL` is **not** set → ask the user once:
  > "I need your Loki base URL to query logs (e.g. `http://loki:3100`). You can also set `LOKI_URL` in your environment to skip this question in future."
- If `LOKI_TOKEN` is not set and the Loki instance requires auth → ask the user once:
  > "Does your Loki require a bearer token? If so please provide it, or set `LOKI_TOKEN` in your environment."

Do **not** ask for both at the same time if the URL is already known. Ask for the token only if the first query returns a 401.

## Step 1 – Detect OS and architecture

```bash
uname -s && uname -m
```

Map to the correct asset name:

| OS / arch             | Asset                          |
|-----------------------|--------------------------------|
| Linux x86_64          | `query-loki-linux-amd64`       |
| Linux aarch64 / arm64 | `query-loki-linux-arm64`       |
| Darwin x86_64         | `query-loki-darwin-amd64`      |
| Darwin arm64          | `query-loki-darwin-arm64`      |
| Windows x86_64        | `query-loki-windows-amd64.exe` |

## Step 2 – Download the binary (if not already present)

The binary lives in `bin/` relative to this skill's directory. Check first:

```bash
ls bin/<asset-name>
```

If missing, download from the latest GitHub Release:

```bash
curl -fsSL -o bin/<asset-name> \
  "https://github.com/vinr-eu/skills/releases/latest/download/<asset-name>"
chmod +x bin/<asset-name>
```

> **Windows**: use `Invoke-WebRequest` or `curl.exe` and skip `chmod`.

## Step 3 – Run the query

With env vars set (preferred):
```bash
./bin/<asset-name> --query '<logql>' [--from <time>] [--to <time>] [--limit N]
```

With explicit flags (fallback):
```bash
./bin/<asset-name> \
  --url <loki-url> \
  --query '<logql>' \
  [--from <RFC3339-or-unix-ns>] \
  [--to   <RFC3339-or-unix-ns>] \
  [--limit 200] \
  [--token <bearer-token>]
```

**Examples:**

Last hour of errors from the nginx job (env vars set):
```bash
./bin/query-loki-linux-amd64 --query '{job="nginx"} |= "error"'
```

Specific time window:
```bash
./bin/query-loki-linux-amd64 \
  --query '{app="api"} | json | level="error"' \
  --from 2024-06-01T08:00:00Z \
  --to   2024-06-01T09:00:00Z \
  --limit 500
```

Raw JSON (for further processing):
```bash
./bin/query-loki-linux-amd64 --query '{job="app"}' --json
```

## Step 4 – Interpret the output

Default mode prints one line per log entry:

```
2024-06-01T08:00:01Z  ERROR something went wrong
2024-06-01T08:00:02Z  ERROR another failure
```

Analyse these lines to answer the user's question — look for patterns, error counts, or specific messages.

## Common issues

- **401 / 403**: token missing or wrong — ask the user for their `LOKI_TOKEN`.
- **No results**: time range too narrow, or LogQL selector matches no streams — widen `--from`/`--to` or relax the query.
- **Binary not executable**: run `chmod +x bin/<asset-name>`.
- **SSL errors**: pass `-k` to curl when downloading from a self-signed instance, or use `--url http://...` instead of https.
