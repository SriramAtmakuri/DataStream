package clickhouse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	host   string
	port   string
	db     string
	user   string
	pass   string
	http   *http.Client
}

type QueryResponse struct {
	Data []map[string]interface{} `json:"data"`
	Rows int                      `json:"rows"`
}

func NewClient() *Client {
	return &Client{
		host: getEnv("CLICKHOUSE_HOST", "localhost"),
		port: getEnv("CLICKHOUSE_PORT", "8123"),
		db:   getEnv("CLICKHOUSE_DB", "ecommerce"),
		user: getEnv("CLICKHOUSE_USER", "default"),
		pass: getEnv("CLICKHOUSE_PASSWORD", ""),
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Query(query string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("http://%s:%s/", c.host, c.port)

	params := url.Values{}
	params.Set("query", query+" FORMAT JSON")
	params.Set("database", c.db)
	params.Set("user", c.user)
	if c.pass != "" {
		params.Set("password", c.pass)
	}

	resp, err := c.http.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("clickhouse http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickhouse error %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return result.Data, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
