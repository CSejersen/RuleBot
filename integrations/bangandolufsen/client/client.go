package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Client struct {
	Config Config
	client *http.Client
	Logger *zap.Logger
}

func New(ctx context.Context, configPath string, logger *zap.Logger) (*Client, error) {
	c := &Client{
		client: &http.Client{},
		Logger: logger,
	}

	err := c.init(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	return c, nil
}

func (c *Client) init(ctx context.Context, configPath string) error {
	err := c.loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config %w", err)
	}

	go func() {
		c.WatchConfig(ctx, configPath)
	}()

	return nil
}

func (c *Client) get(ip, path string, v interface{}) error {
	return c.doRequest(http.MethodGet, ip, path, nil, v)
}

func (c *Client) post(ip, path string, body, v interface{}) error {
	return c.doRequest(http.MethodPost, ip, path, body, v)
}

func (c *Client) put(ip, path string, body, v interface{}) error {
	return c.doRequest(http.MethodPut, ip, path, body, v)
}

func (c *Client) doRequest(method, ip, path string, body interface{}, v interface{}) error {
	url := fmt.Sprintf("http://%s/api/v1/%s", ip, path)

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	//c.Logger.Debug("sending request",
	//	zap.String("method", method),
	//	zap.String("url", url),
	//	zap.Any("body", body))

	resp, err := c.client.Do(req)
	if err != nil {
		c.Logger.Error("request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %s\n%s", resp.Status, string(body))
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}
