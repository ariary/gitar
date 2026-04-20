package webhook

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ariary/gitar/pkg/config"
)

func TestProcessRequestBodyRestoredAfterFullBodyRead(t *testing.T) {
	body := `{"key":"value","secret":"abc"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	cfg := &config.ConfigWebHook{FullBody: true}
	ProcessRequest(req, cfg)

	// Body must still be readable by downstream handlers
	remaining, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body after ProcessRequest: %v", err)
	}
	if string(remaining) != body {
		t.Errorf("body not restored: got %q, want %q", string(remaining), body)
	}
}

func TestProcessRequestBodyUnchangedWhenFullBodyFalse(t *testing.T) {
	body := `hello=world`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg := &config.ConfigWebHook{FullBody: false}
	ProcessRequest(req, cfg)

	remaining, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	// Body not consumed when FullBody is false
	if string(remaining) != body {
		t.Errorf("body unexpectedly consumed: got %q, want %q", string(remaining), body)
	}
}
