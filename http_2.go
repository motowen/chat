package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// HttpClient 定義抽象介面
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
			time.Sleep(time.Second * 1) // 簡單的 retry backoff
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

// ---------------- Main ----------------
func main() {
	url := "https://httpbin.org/post"
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte(`{"name":"moto","role":"backend"}`)

	// 1. 使用 net/http (timeout 5s, retry 2)
	var client HttpClient = NewNetHttpClient(5*time.Second, 2)
	resp, err := client.DoPost(url, headers, body)
	if err != nil {
		panic(err)
	}
	fmt.Println("[net/http] Response:", resp)

	// 2. 使用 resty (timeout 5s, retry 2)
	client = NewRestyClient(5*time.Second, 2)
	resp, err = client.DoPost(url, headers, body)
	if err != nil {
		panic(err)
	}
	fmt.Println("[resty] Response:", resp)
}
