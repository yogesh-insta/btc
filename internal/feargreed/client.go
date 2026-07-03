package feargreed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultURL = "https://api.alternative.me/fng/?limit=1"

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type response struct {
	Data []struct {
		Value string `json:"value"`
	} `json:"data"`
}

func (c *Client) Fetch(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	if res.StatusCode >= 300 {
		return 0, fmt.Errorf("fear & greed API HTTP %d", res.StatusCode)
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, fmt.Errorf("decode fear & greed response: %w", err)
	}
	if len(parsed.Data) == 0 || parsed.Data[0].Value == "" {
		return 0, fmt.Errorf("fear & greed API returned no value")
	}

	value, err := strconv.Atoi(parsed.Data[0].Value)
	if err != nil {
		return 0, fmt.Errorf("parse fear & greed value %q: %w", parsed.Data[0].Value, err)
	}
	return value, nil
}
