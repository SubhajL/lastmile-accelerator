package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type promResp struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

type promVectorResult []struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"`
}

func TestPromClient_Query_Vector_Success(t *testing.T) {
	// Arrange a fake Prometheus server returning vector
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":["%d","123.45"]}]}}`, time.Now().Unix())
	}))
	defer srv.Close()

	c := NewPromClient(srv.URL, &http.Client{Timeout: 2 * time.Second})
	val, err := c.Query(context.Background(), "up", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 123.45 {
		t.Fatalf("want 123.45 got %v", val)
	}
}

func TestPromClient_Query_Scalar_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","data":{"resultType":"scalar","result":["%d","42"]}}`, time.Now().Unix())
	}))
	defer srv.Close()

	c := NewPromClient(srv.URL, nil)
	val, err := c.Query(context.Background(), "up", time.Minute)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if val != 42 {
		t.Fatalf("want 42 got %v", val)
	}
}

func TestPromClient_Query_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer srv.Close()

	c := NewPromClient(srv.URL, nil)
	_, err := c.Query(context.Background(), "up", time.Minute)
	if err == nil {
		t.Fatal("expected error on HTTP 500")
	}
}

func TestPromClient_Query_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := NewPromClient(srv.URL, nil)
	_, err := c.Query(context.Background(), "up", time.Minute)
	if err == nil {
		t.Fatal("expected json parse error")
	}
}

func TestPromClient_Query_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer srv.Close()

	c := NewPromClient(srv.URL, nil)
	_, err := c.Query(context.Background(), "up", time.Minute)
	if err == nil {
		t.Fatal("expected empty result error")
	}
}
