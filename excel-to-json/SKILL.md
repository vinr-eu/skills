---
name: excel-to-json
description: Converts an Excel (.xlsx/.xls) file into a JSON array for analysis. Use when the user provides an Excel file path and wants to inspect, query, or analyze its contents. Accepts an optional sheet name (defaults to first sheet) and optional output file path. If an output file is specified the JSON is written to disk and not echoed; otherwise the full JSON array is returned for the agent to parse and reason over.
compatibility: Requires Node.js (auto-installed via nvm if missing). The xlsx npm package is auto-installed on first run.
metadata:
  author: vinr
  version: "1.0"
---

## Overview

This skill reads an Excel file and converts a worksheet into a JSON array, where each element represents one row as a `{ "column": value }` object. This makes the data fully legible to an agent for filtering, summarising, pivoting, or answering user questions.

## Parameters

| Parameter | Flag | Required | Default | Notes |
|---|---|---|---|---|
| Excel file path | positional | Yes | — | Absolute or relative path to `.xlsx` or `.xls` |
| Sheet name | `--sheet` / `-s` | No | First sheet | Exact sheet name (case-sensitive) |
| Output file | `--output` / `-o` | No | None (stdout) | If given, write JSON to this path instead of printing |

## Step 1 – Ensure Node.js is available

Check if Node is installed:

```bash
node --version
```

If the command is not found, install Node via nvm:

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
source ~/.nvm/nvm.sh
nvm install --lts
```

Or via a system package manager:

```bash
# macOS
brew install node

# Debian / Ubuntu
sudo apt-get install -y nodejs npm

# Fedora / RHEL
sudo dnf install -y nodejs
```

## Step 2 – Run the conversion script

The `xlsx` npm package is installed automatically on first run — no manual `npm install` needed.

```bash
node scripts/convert.js <excel_file> [--sheet <sheet_name>] [--output <output_file>]
```

**Examples:**

Convert the first sheet, print JSON to stdout (agent reads it directly):
```bash
node scripts/convert.js /path/to/data.xlsx
```

Convert a specific sheet:
```bash
node scripts/convert.js /path/to/data.xlsx --sheet "Sales Q1"
```

Convert and save to a file (keeps context window clean for large files):
```bash
node scripts/convert.js /path/to/data.xlsx --sheet "Inventory" --output /tmp/inventory.json
```

## Step 3 – Analyse the output

When **no output file** is specified the script prints the JSON array to stdout. Parse it and reason over the records to answer the user's question.

When an **output file** is specified the script prints a one-line confirmation (`JSON written to <path> (<N> records)`). Read the file with the Read tool when needed.

## Output format

Each row becomes a JSON object. Column headers become keys. Empty cells become `null`. Dates are ISO 8601 strings. Numbers keep their native type.

```json
[
  { "Name": "Alice", "Age": 30, "Score": 95.5,  "Joined": "2023-01-15" },
  { "Name": "Bob",   "Age": null, "Score": 88.0, "Joined": "2022-07-01" }
]
```

## Discovering sheet names

If the user has not specified a sheet, or the requested sheet is not found, the script prints all available sheet names in the error message. You can also list them explicitly:

```bash
node -e "const X=require('./scripts/node_modules/xlsx'); const wb=X.readFile(process.argv[1]); console.log(wb.SheetNames);" /path/to/data.xlsx
```

## Common edge cases

- **Large files (>50 k rows)**: always use `--output` to avoid flooding the context window, then query the saved JSON file.
- **Merged cells**: SheetJS fills merged regions with the top-left cell value.
- **Multiple header rows**: the first row is treated as headers by default. If the file has a different layout, note the row index and adjust manually.
- **Formula cells**: SheetJS reads cached values. Warn the user if values look stale and the file hasn't been opened recently.
- **CSV or other formats**: this skill targets `.xlsx`/`.xls` only; redirect the user to use standard tools for CSV.
