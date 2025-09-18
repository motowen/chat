func TestNetHttpClient_Retry(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	callCount := 0

	httpmock.RegisterResponder("POST", "https://example.com/retry",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				// 第一次回傳錯誤，模擬 timeout
				return nil, fmt.Errorf("timeout")
			}
			// 第二次成功
			return httpmock.NewStringResponse(200, `{"status":"ok","via":"net/http"}`), nil
		})

	client := NewNetHttpClient(2*time.Second, 2)
	resp, err := client.DoPost("https://example.com/retry",
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"name":"moto"}`),
	)

	assert.NoError(t, err)
	assert.Contains(t, resp, `"status":"ok"`)
	assert.Equal(t, 2, callCount, "應該重試 1 次後成功")
}

func TestRestyClient_Retry(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	callCount := 0

	httpmock.RegisterResponder("POST", "https://example.com/retry",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 3 {
				// 前兩次失敗
				return nil, fmt.Errorf("temporary network error")
			}
			// 第三次成功
			return httpmock.NewStringResponse(200, `{"status":"ok","via":"resty"}`), nil
		})

	// resty retryCount=2，代表最多會呼叫 1 + 2 = 3 次
	client := NewRestyClient(2*time.Second, 2)
	resp, err := client.DoPost("https://example.com/retry",
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"name":"moto"}`),
	)

	assert.NoError(t, err)
	assert.Contains(t, resp, `"status":"ok"`)
	assert.Equal(t, 3, callCount, "應該重試 2 次後成功")
}
