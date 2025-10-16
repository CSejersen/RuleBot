package eventsource

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type EventSource struct {
	IP     string
	AppKey string
	Logger *zap.Logger
}

func New(ip string, appKey string, logger *zap.Logger) *EventSource {
	return &EventSource{
		IP:     ip,
		AppKey: appKey,
		Logger: logger,
	}
}

func (s *EventSource) Run(ctx context.Context, out chan<- []byte) error {
	for {
		err := s.connectAndStream(ctx, out)
		if err != nil {
			s.Logger.Error("hue event stream disconnected", zap.Error(err))
		} else {
			s.Logger.Info("hue event stream closed gracefully")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 3):
			continue // attempt to reconnect
		}
	}
}

func (s *EventSource) connectAndStream(ctx context.Context, out chan<- []byte) error {
	url := fmt.Sprintf("https://%s/eventstream/clip/v2", s.IP)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("hue-application-key", s.AppKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore bridge self-signed cert
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	s.Logger.Info("hue event stream connected", zap.String("url", url))

	scanner := bufio.NewScanner(resp.Body)
	var buffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if buffer.Len() > 0 {
				out <- buffer.Bytes()
				buffer.Reset()
			}
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			buffer.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}

	return scanner.Err()
}
