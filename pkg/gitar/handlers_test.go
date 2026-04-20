package gitar

import (
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/ariary/gitar/pkg/config"
)

func aliasScript(t *testing.T) string {
	t.Helper()
	cfg := &config.Config{
		Url:         "http://127.0.0.1:9292/testsecret",
		DownloadDir: t.TempDir(),
		Completion:  false,
	}
	req := httptest.NewRequest("GET", "/testsecret/alias", nil)
	w := httptest.NewRecorder()
	AliasHandler(cfg)(w, req)
	return w.Body.String()
}

func TestAliasScriptBashSyntax(t *testing.T) {
	script := aliasScript(t)
	cmd := exec.Command("bash", "-n")
	cmd.Stdin = strings.NewReader(script)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("alias script has bash syntax error: %v\n%s\nScript:\n%s", err, out, script)
	}
}

func TestAliasScriptZshSyntax(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not available")
	}
	script := aliasScript(t)
	cmd := exec.Command("zsh", "-n")
	cmd.Stdin = strings.NewReader(script)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("alias script has zsh syntax error: %v\n%s\nScript:\n%s", err, out, script)
	}
}

func TestAliasWindowsPSContainsPullrPushr(t *testing.T) {
	cfg := &config.Config{
		Url:         "http://127.0.0.1:9292/testsecret",
		DownloadDir: t.TempDir(),
		Completion:  false,
	}
	req := httptest.NewRequest("GET", "/testsecret/aliaswinps", nil)
	w := httptest.NewRecorder()
	AliasWindowsPS(cfg)(w, req)
	script := w.Body.String()

	for _, fn := range []string{"function pullr", "function pushr"} {
		if !strings.Contains(script, fn) {
			t.Errorf("Windows PS alias missing %q", fn)
		}
	}
	if !strings.Contains(script, "pushrzip") {
		t.Error("Windows PS pushr alias does not reference /pushrzip endpoint")
	}
}

func TestAliasScriptContainsPullr(t *testing.T) {
	script := aliasScript(t)
	if !strings.Contains(script, "pullr()") {
		t.Error("alias script missing pullr function")
	}
	if !strings.Contains(script, "pushr()") {
		t.Error("alias script missing pushr function")
	}
}
