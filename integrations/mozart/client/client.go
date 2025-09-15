package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type MozartDevice struct {
	IP  string
	JID string
}

type Client struct {
	DeviceRegistry map[string]MozartDevice // humanID --> MozartDevice
	client         *http.Client
	Logger         *zap.Logger
}

func New(deviceIPs map[string]string, logger *zap.Logger) (*Client, error) {
	c := &Client{
		client: &http.Client{},
		Logger: logger,
	}

	if err := c.initDeviceRegistry(deviceIPs); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) initDeviceRegistry(deviceIPs map[string]string) error {
	c.DeviceRegistry = make(map[string]MozartDevice)

	for name, ip := range deviceIPs {
		jid, err := c.fetchJID(ip)
		if err != nil {
			return err
		}

		c.DeviceRegistry[name] = MozartDevice{
			IP:  ip,
			JID: jid,
		}
	}
	return nil
}

func (c *Client) get(deviceIP, path string, v interface{}) error {
	return c.doRequest(http.MethodGet, deviceIP, path, nil, v)
}

func (c *Client) post(deviceIP, path string, body, v interface{}) error {
	return c.doRequest(http.MethodPost, deviceIP, path, body, v)
}

func (c *Client) put(deviceIP, path string, body, v interface{}) error {
	return c.doRequest(http.MethodPut, deviceIP, path, body, v)
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
