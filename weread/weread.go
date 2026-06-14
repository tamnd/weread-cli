// Package weread is the library behind the weread command: the HTTP client,
// request shaping, and the typed data models for WeRead (weread.qq.com).
//
// The client GETs the public search JSON endpoint with the required User-Agent
// and Referer headers, paces requests, and retries transient 429/5xx errors
// with exponential backoff. No authentication is required.
package weread

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent is the browser UA WeRead expects.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Referer   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://weread.qq.com",
		UserAgent: DefaultUserAgent,
		Referer:   "https://weread.qq.com/",
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the WeRead JSON API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Search queries the WeRead global search endpoint and returns up to limit books.
// If limit <= 0 the full result set is returned.
func (c *Client) Search(ctx context.Context, query string, limit int) ([]Book, error) {
	u := c.cfg.BaseURL + "/web/search/global?keyword=" + url.QueryEscape(query)

	raw, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}

	n := len(resp.Books)
	if limit > 0 && limit < n {
		n = limit
	}

	out := make([]Book, 0, n)
	for i, entry := range resp.Books[:n] {
		info := entry.BookInfo
		bookURL := ""
		if info.BookID != "" {
			bookURL = c.cfg.BaseURL + "/web/bookDetail/" + info.BookID
		}
		out = append(out, Book{
			Rank:   i + 1,
			Title:  strings.TrimSpace(info.Title),
			Author: strings.TrimSpace(info.Author),
			BookID: info.BookID,
			URL:    bookURL,
		})
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, u)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", u, lastErr)
}

func (c *Client) do(ctx context.Context, u string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Referer", c.cfg.Referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
