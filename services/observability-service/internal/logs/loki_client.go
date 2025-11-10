package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	base string
	hc   *http.Client
}

func NewClient(base string, hc *http.Client) *Client {
	if hc == nil { hc = &http.Client{Timeout: 10 * time.Second} }
	return &Client{base: base, hc: hc}
}

type LogEntry struct {
	Timestamp time.Time         `json:"ts"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// QueryRange executes a Loki range query for logs and flattens into entries.
func (c *Client) QueryRange(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]LogEntry, error) {
	u, _ := url.Parse(c.base)
	u.Path = "/loki/api/v1/query_range"
	qs := u.Query()
	qs.Set("query", q)
	if !start.IsZero() { qs.Set("start", strconv.FormatInt(start.UnixNano(), 10)) }
	if !end.IsZero() { qs.Set("end", strconv.FormatInt(end.UnixNano(), 10)) }
	if limit > 0 { qs.Set("limit", strconv.Itoa(limit)) }
	if direction != "" { qs.Set("direction", direction) }
	u.RawQuery = qs.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.hc.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("loki http %d", resp.StatusCode) }
	var lr struct {
		Status string `json:"status"`
		Data struct {
			ResultType string `json:"resultType"`
			Result []struct {
				Stream map[string]string `json:"stream"`
				Values [][2]string       `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil { return nil, err }
	if lr.Status != "success" { return nil, fmt.Errorf("loki status %s", lr.Status) }
	var entries []LogEntry
	for _, s := range lr.Data.Result {
		for _, v := range s.Values {
			ns, _ := strconv.ParseInt(v[0], 10, 64)
			entries = append(entries, LogEntry{Timestamp: time.Unix(0, ns), Line: v[1], Labels: s.Stream})
		}
	}
	return entries, nil
}