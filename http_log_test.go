package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	// 建立一個假的 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"message":"ok"}`)
	})

	// 包上 logging middleware
	ts := httptest.NewServer(LoggingMiddleware(handler))
	defer ts.Close()

	// 捕捉 log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(io.Discard) // 測試後丟掉

	// 發送請求
	resp, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	// 驗證 response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// 驗證 log output
	output := logBuf.String()
	if !strings.Contains(output, "[REQUEST] GET /test") {
		t.Errorf("expected request log, got: %s", output)
	}
	if !strings.Contains(output, "[RESPONSE] status=200 body={\"message\":\"ok\"}") {
		t.Errorf("expected response log, got: %s", output)
	}
}
