package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type HueType string

type Client struct {
	IP             string
	AppKey         string
	DeviceRegistry DeviceRegistry // human-readableID -> hueID
	client         *http.Client
	Logger         *zap.Logger
}

func New(ip string, appKey string, logger *zap.Logger) (*Client, error) {
	c := &Client{
		IP:     ip,
		AppKey: appKey,
		client: newHTTPClient(),
		Logger: logger,
	}

	if err := c.InitRegistry(); err != nil {
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

func (c *Client) get(path string, v interface{}) error {
	return c.doApiV2Request(http.MethodGet, path, nil, v)
}

func (c *Client) post(path string, body, v interface{}) error {
	return c.doApiV2Request(http.MethodPost, path, body, v)
}

func (c *Client) put(path string, body, v interface{}) error {
	return c.doApiV2Request(http.MethodPut, path, body, v)
}

func (c *Client) doApiV2Request(method, path string, body interface{}, v interface{}) error {
	url := fmt.Sprintf("https://%s/clip/v2/%s", c.IP, path)

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

	if c.AppKey == "" {
		return errors.New("missing AppKey")
	}
	req.Header.Set("hue-application-key", c.AppKey)

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
