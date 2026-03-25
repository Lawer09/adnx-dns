package godaddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ErrRateLimited struct{ Message string }
func (e *ErrRateLimited) Error() string { return e.Message }

type DomainSummary struct {
	Domain string `json:"domain"`
}

type Record struct {
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`
}

type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	limit      int
	mu         sync.Mutex
	window     time.Time
	count      int
}

func NewClient(baseURL, key, secret string, timeoutSec, limit int) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey:key, apiSecret:secret, limit:limit,
		httpClient:&http.Client{Timeout: time.Duration(timeoutSec) * time.Second}, window: time.Now()}
}

func (c *Client) allow() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if now.Sub(c.window) >= time.Minute {
		c.window = now
		c.count = 0
	}
	if c.count >= c.limit {
		return &ErrRateLimited{Message: fmt.Sprintf("godaddy local rate limit exceeded: %d requests/minute", c.limit)}
	}
	c.count++
	return nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	if err := c.allow(); err != nil { return err }
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil { return err }
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil { return err }
	req.Header.Set("Authorization", "sso-key "+c.apiKey+":"+c.apiSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusTooManyRequests {
		return &ErrRateLimited{Message: fmt.Sprintf("godaddy responded with 429: %s", string(respBody))}
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("godaddy api error status=%d body=%s", resp.StatusCode, string(respBody))
	}
	if out != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (c *Client) ListDomains(ctx context.Context) ([]DomainSummary, error) {
	var out []DomainSummary
	err := c.do(ctx, http.MethodGet, "/v1/domains", nil, &out)
	return out, err
}

func (c *Client) UpsertARecord(ctx context.Context, domain, subdomain, ip string, ttl int) error {
	path := fmt.Sprintf("/v1/domains/%s/records/A/%s", domain, subdomain)
	return c.do(ctx, http.MethodPut, path, []Record{{Data: ip, TTL: ttl}}, nil)
}

func (c *Client) DeleteARecord(ctx context.Context, domain, subdomain string) error {
	path := fmt.Sprintf("/v1/domains/%s/records/A/%s", domain, subdomain)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
