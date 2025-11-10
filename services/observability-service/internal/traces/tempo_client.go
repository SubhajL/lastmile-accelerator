package traces

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

type TraceSummary struct {
	TraceID    string `json:"traceID"`
	Service    string `json:"service"`
	Operation  string `json:"operation"`
	DurationMs int64  `json:"durationMs"`
	Error      bool   `json:"error"`
}

type TraceDetail struct {
	TraceID string            `json:"traceID"`
	Spans   []json.RawMessage `json:"spans"`
}

// SearchTraceQL issues a Tempo TraceQL search against /api/search/traceql
func (c *Client) SearchTraceQL(ctx context.Context, query string, start, end time.Time, limit int) ([]TraceSummary, error) {
	u, _ := url.Parse(c.base)
	u.Path = "/api/search/traceql"
	q := u.Query()
	q.Set("query", query)
	if !start.IsZero() { q.Set("start", strconv.FormatInt(start.UnixNano(), 10)) }
	if !end.IsZero() { q.Set("end", strconv.FormatInt(end.UnixNano(), 10)) }
	if limit > 0 { q.Set("limit", strconv.Itoa(limit)) }
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.hc.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("tempo http %d", resp.StatusCode) }
	var out []TraceSummary
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
	return out, nil
}

func (c *Client) GetTrace(ctx context.Context, traceID string) (*TraceDetail, error) {
	u, _ := url.Parse(c.base)
	u.Path = "/api/traces/" + traceID
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.hc.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("tempo http %d", resp.StatusCode) }
	var detail TraceDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil { return nil, err }
	return &detail, nil
}
