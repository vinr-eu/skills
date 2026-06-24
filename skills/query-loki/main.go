package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type queryResponse struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][2]string       `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	lokiURL := flag.String("url", os.Getenv("LOKI_URL"), "Loki base URL (env: LOKI_URL)")
	query := flag.String("query", "", "LogQL query (e.g. {job=\"nginx\"} |= \"error\")")
	from := flag.String("from", "", "Start time (RFC3339 or Unix ns, default: 1 hour ago)")
	to := flag.String("to", "", "End time (RFC3339 or Unix ns, default: now)")
	limit := flag.Int("limit", 100, "Max log lines to return")
	token := flag.String("token", os.Getenv("LOKI_TOKEN"), "Bearer token for auth (env: LOKI_TOKEN)")
	jsonOut := flag.Bool("json", false, "Print raw JSON response instead of plain log lines")
	flag.Parse()

	if *lokiURL == "" {
		slog.Error("Loki URL required", "hint", "set LOKI_URL env var or pass --url")
		os.Exit(1)
	}
	if *query == "" {
		slog.Error("query required", "hint", "pass --query '<logql>'")
		os.Exit(1)
	}

	now := time.Now()
	startNs := strconv.FormatInt(now.Add(-1*time.Hour).UnixNano(), 10)
	endNs := strconv.FormatInt(now.UnixNano(), 10)

	if *from != "" {
		startNs = parseTime(*from)
	}
	if *to != "" {
		endNs = parseTime(*to)
	}

	params := url.Values{}
	params.Set("query", *query)
	params.Set("start", startNs)
	params.Set("end", endNs)
	params.Set("limit", strconv.Itoa(*limit))
	params.Set("direction", "forward")

	req, err := http.NewRequest("GET", *lokiURL+"/loki/api/v1/query_range?"+params.Encode(), nil)
	if err != nil {
		slog.Error("error building request", "error", err)
		os.Exit(1)
	}
	if *token != "" {
		req.Header.Set("Authorization", "Bearer "+*token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error querying Loki", "url", *lokiURL, "error", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response body", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected response from Loki", "status", resp.StatusCode, "body", string(body))
		os.Exit(1)
	}

	if *jsonOut {
		fmt.Println(string(body))
		return
	}

	var result queryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("error parsing response", "error", err)
		os.Exit(1)
	}

	for _, stream := range result.Data.Result {
		for _, v := range stream.Values {
			ts, _ := strconv.ParseInt(v[0], 10, 64)
			t := time.Unix(0, ts).UTC().Format(time.RFC3339)
			fmt.Printf("%s  %s\n", t, v[1])
		}
	}
}

func parseTime(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return strconv.FormatInt(t.UnixNano(), 10)
	}
	return s
}
