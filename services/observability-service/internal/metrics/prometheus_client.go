package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PromClient struct {
	baseURL string
	hc      *http.Client
}

func NewPromClient(baseURL string, httpClient *http.Client) *PromClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &PromClient{baseURL: baseURL, hc: httpClient}
}

// Query executes a Prometheus instant query and returns a numeric value.
// It supports scalar and vector (first sample) responses.
func (p *PromClient) Query(ctx context.Context, promQL string, window time.Duration) (float64, error) {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return 0, fmt.Errorf("invalid base url: %w", err)
	}
	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", promQL)
	q.Set("time", strconv.FormatInt(time.Now().Unix(), 10))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	resp, err := p.hc.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus http %d", resp.StatusCode)
	}
	var pr struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}
	if pr.Status != "success" {
		return 0, fmt.Errorf("prometheus status: %s", pr.Status)
	}
	switch pr.Data.ResultType {
	case "scalar":
		var arr [2]any
		if err := json.Unmarshal(pr.Data.Result, &arr); err != nil {
			return 0, fmt.Errorf("scalar unmarshal: %w", err)
		}
		if len(arr) != 2 {
			return 0, fmt.Errorf("scalar unexpected format")
		}
		vs, ok := arr[1].(string)
		if !ok {
			return 0, fmt.Errorf("scalar value not string")
		}
		vf, err := strconv.ParseFloat(vs, 64)
		if err != nil {
			return 0, fmt.Errorf("scalar parse: %w", err)
		}
		return vf, nil
	case "vector":
		var vec []struct {
			Value [2]any `json:"value"`
		}
		if err := json.Unmarshal(pr.Data.Result, &vec); err != nil {
			return 0, fmt.Errorf("vector unmarshal: %w", err)
		}
		if len(vec) == 0 {
			return 0, fmt.Errorf("empty result")
		}
		vs, ok := vec[0].Value[1].(string)
		if !ok {
			return 0, fmt.Errorf("vector value not string")
		}
		vf, err := strconv.ParseFloat(vs, 64)
		if err != nil {
			return 0, fmt.Errorf("vector parse: %w", err)
		}
		return vf, nil
	default:
		return 0, fmt.Errorf("unsupported resultType: %s", pr.Data.ResultType)
	}
}
