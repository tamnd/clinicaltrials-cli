// Package clinicaltrials is the library behind the clinicaltrials command: the
// HTTP client, request shaping, and the typed data models for the
// ClinicalTrials.gov REST API v2.
//
// The public API at https://clinicaltrials.gov/api/v2 is fully open — no API
// key, no authentication required. The client sets a real User-Agent, paces
// requests at 500 ms by default, and retries 429/5xx with exponential backoff.
package clinicaltrials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Host is the ClinicalTrials.gov hostname.
const Host = "clinicaltrials.gov"

// DefaultUserAgent identifies the client to ClinicalTrials.gov.
const DefaultUserAgent = "clinicaltrials/dev (+https://github.com/tamnd/clinicaltrials-cli)"

// ErrNotFound is returned when the API returns a 404 for an NCT ID.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://" + Host,
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the ClinicalTrials.gov REST API v2.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
	rate       time.Duration
	retries    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	base := cfg.BaseURL
	if base == "" {
		base = "https://" + Host
	}
	return &Client{
		baseURL:    base,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
	}
}

// Search returns studies matching condition, intervention, term, and/or status.
// limit 0 uses the default of 10.
func (c *Client) Search(ctx context.Context, condition, intervention, term, status string, limit int) ([]Study, error) {
	params := url.Values{"format": {"json"}}
	if condition != "" {
		params.Set("query.cond", condition)
	}
	if intervention != "" {
		params.Set("query.intr", intervention)
	}
	if term != "" {
		params.Set("query.term", term)
	}
	if status != "" {
		params.Set("filter.overallStatus", status)
	}
	return c.collectStudies(ctx, params, effectiveLimit(limit, 10))
}

// GetStudy returns a single study by NCT ID.
func (c *Client) GetStudy(ctx context.Context, nctID string) (*Study, error) {
	nctID = normalizeNCT(nctID)
	rawURL := c.baseURL + "/api/v2/studies/" + url.PathEscape(nctID) + "?format=json"
	var s wireStudy
	if err := c.getJSON(ctx, rawURL, &s); err != nil {
		return nil, fmt.Errorf("study %s: %w", nctID, err)
	}
	study := toStudy(s)
	return &study, nil
}

// ─── pagination ──────────────────────────────────────────────────────────────

func (c *Client) collectStudies(ctx context.Context, params url.Values, limit int) ([]Study, error) {
	pageSize := limit
	if pageSize > 100 {
		pageSize = 100
	}
	params.Set("pageSize", strconv.Itoa(pageSize))

	var studies []Study
	token := ""
	for {
		page := url.Values{}
		for k, v := range params {
			page[k] = v
		}
		if token != "" {
			page.Set("pageToken", token)
		}

		rawURL := c.baseURL + "/api/v2/studies?" + page.Encode()
		var resp wireListResponse
		if err := c.getJSON(ctx, rawURL, &resp); err != nil {
			return studies, err
		}
		for _, s := range resp.Studies {
			studies = append(studies, toStudy(s))
			if len(studies) >= limit {
				return studies, nil
			}
		}
		if resp.NextPageToken == "" || len(resp.Studies) == 0 {
			break
		}
		token = resp.NextPageToken
	}
	return studies, nil
}

// ─── HTTP ─────────────────────────────────────────────────────────────────────

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

func effectiveLimit(n, def int) int {
	if n > 0 {
		return n
	}
	return def
}

// normalizeNCT uppercases and ensures the NCT prefix.
func normalizeNCT(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	if !strings.HasPrefix(id, "NCT") {
		id = "NCT" + id
	}
	return id
}
