package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

// HttpClient 抽象介面
type HttpClient interface {
	DoPost(url string, headers map[string]string, body []byte) (string, error)
}

// ---------------- net/http 實作 ----------------
type NetHttpClient struct {
	client     *http.Client
	retryCount int
}

func NewNetHttpClient(timeout time.Duration, retryCount int) *NetHttpClient {
	return &NetHttpClient{
		client:     &http.Client{Timeout: timeout},
		retryCount: retryCount,
	}
}

func (c *NetHttpClient) DoPost(url string, headers map[string]string, body []byte) (string, error) {
	var lastErr error
	for i := 0; i <= c.retryCount; i++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return "", err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * 1)
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(respBody), nil
	}
	return "", lastErr
}

// ---------------- resty 實作 ----------------
type RestyClient struct {
	client *resty.Client
}

func NewRestyClient(timeout time.Duration, retryCount int) *RestyClient {
	client := resty.New()
	client.SetTimeout(timeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(time.Second * 1).
		SetRetryMaxWaitTime(time.Second * 5)

	return &RestyClient{client: client}
}

func (c *RestyClient) DoPost(url string, headers map[string]string, body []byte) (string, error) {
	req := c.client.R().SetBody(body)
	for k, v := range headers {
		req.SetHeader(k, v)
	}

	resp, err := req.Post(url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

// ---------------- Logging Decorator ----------------
type LoggingHttpClient struct {
	inner HttpClient
	logger *slog.Logger
}

func NewLoggingHttpClient(inner HttpClient) *LoggingHttpClient {
	return &LoggingHttpClient{
		inner:  inner,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (c *LoggingHttpClient) DoPost(url string, headers map[string]string, body []byte) (string, error) {
	c.logger.Info("HTTP POST Request",
		"url", url,
		"headers", headers,
		"body", string(body),
	)

	start := time.Now()
	resp, err := c.inner.DoPost(url, headers, body)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Error("HTTP POST Failed", "url", url, "error", err, "elapsed", elapsed)
		return "", err
	}

	c.logger.Info("HTTP POST Response",
		"url", url,
		"elapsed", elapsed,
		"resp", resp,
	)
	return resp, nil
}

// ---------------- Main ----------------
func main() {
	url := "https://httpbin.org/post"
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte(`{"name":"moto","role":"backend"}`)

	// 1. 使用 net/http + logging
	netClient := NewNetHttpClient(5*time.Second, 2)
	client := NewLoggingHttpClient(netClient)
	client.DoPost(url, headers, body)

	// 2. 使用 resty + logging
	restyClient := NewRestyClient(5*time.Second, 2)
	client = NewLoggingHttpClient(restyClient)
	client.DoPost(url, headers, body)
}
