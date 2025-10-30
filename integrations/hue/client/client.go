package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type ApiClient struct {
	IP               string
	AppKey           string
	ResourceRegistry ResourceRegistry
	client           *http.Client
	Logger           *zap.Logger
}

func New(ip string, appKey string, logger *zap.Logger) (*ApiClient, error) {
	c := &ApiClient{
		IP:     ip,
		AppKey: appKey,
		client: newHTTPClient(),
		Logger: logger,
	}

	if err := c.BuildResourceRegistry(); err != nil {
		return nil, err
	}

	return c, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore bridge self-signed cert
		},
	}
}

func (c *ApiClient) get(ctx context.Context, path string, v interface{}) error {
	return c.doApiV2Request(ctx, http.MethodGet, path, nil, v)
}

func (c *ApiClient) post(ctx context.Context, path string, body, v interface{}) error {
	return c.doApiV2Request(ctx, http.MethodPost, path, body, v)
}

func (c *ApiClient) put(ctx context.Context, path string, body, v interface{}) error {
	return c.doApiV2Request(ctx, http.MethodPut, path, body, v)
}

func (c *ApiClient) doApiV2Request(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	url := fmt.Sprintf("https://%s/clip/v2/%s", c.IP, path)

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}

	if c.AppKey == "" {
		return errors.New("missing AppKey")
	}
	req.Header.Set("hue-application-key", c.AppKey)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.Logger.Error("request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if v != nil {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("failed to decode response: %w\nbody: %s", err, string(respBody))
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s\nbody: %s", resp.Status, string(respBody))
	}

	return nil
}
