package main

import (
	"testing"
	"time"

	httpmock "github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNetHttpClient_DoPost(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 模擬一個假的 API 回應
	httpmock.RegisterResponder("POST", "https://example.com/post",
		httpmock.NewStringResponder(200, `{"status":"ok","via":"net/http"}`))

	client := NewNetHttpClient(2*time.Second, 1)
	resp, err := client.DoPost("https://example.com/post",
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"name":"moto"}`),
	)

	assert.NoError(t, err)
	assert.Contains(t, resp, `"status":"ok"`)
	assert.Contains(t, resp, `"via":"net/http"`)
}

func TestRestyClient_DoPost(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 模擬一個假的 API 回應
	httpmock.RegisterResponder("POST", "https://example.com/post",
		httpmock.NewStringResponder(200, `{"status":"ok","via":"resty"}`))

	client := NewRestyClient(2*time.Second, 1)
	resp, err := client.DoPost("https://example.com/post",
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"name":"moto"}`),
	)

	assert.NoError(t, err)
	assert.Contains(t, resp, `"status":"ok"`)
	assert.Contains(t, resp, `"via":"resty"`)
}
