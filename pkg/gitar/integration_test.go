package gitar

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// generateTestCerts creates a self-signed TLS certificate in dir.
func generateTestCerts(t *testing.T, dir string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	if err := os.WriteFile(filepath.Join(dir, "server.crt"), certPEM, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "server.key"), keyPEM, 0600); err != nil {
		t.Fatal(err)
	}
}

// freePort returns a free TCP port on localhost.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return port
}

// buildGitar compiles the gitar binary into tmpDir and returns the path.
func buildGitar(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "gitar")
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = projectRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binPath
}

// startGitar launches the gitar binary and waits for it to be ready.
// Returns a cleanup function that kills the process.
func startGitar(t *testing.T, binPath string, args ...string) (*exec.Cmd, func()) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start gitar: %v", err)
	}
	return cmd, func() { cmd.Process.Kill() }
}

// waitForHTTP polls a URL until it returns 200 or the deadline is reached.
func waitForHTTP(t *testing.T, client *http.Client, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", url)
}

// ──────────────────────────────────────────────
// Issue #13: TLS + port forwarding
// ──────────────────────────────────────────────

func TestIntegrationTLSPortForwarding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)

	// Generate TLS certs
	certDir := filepath.Join(t.TempDir(), "certs")
	os.Mkdir(certDir, 0755)
	generateTestCerts(t, certDir)

	// Start echo TCP server (target service)
	echoLn, echoPort := echoServer(t)
	defer echoLn.Close()

	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "inttest",
		"-p", gitarPort,
		"-f", echoPort,
		"--tls", "-x", certDir,
		"--copy=false",
	)
	defer cleanup()

	// Wait for HTTPS server
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 1 * time.Second,
	}
	waitForHTTP(t, httpClient, "https://127.0.0.1:"+gitarPort+"/inttest/alias", 5*time.Second)

	// Shutdown HTTPS server
	resp, err := httpClient.Get("https://127.0.0.1:" + gitarPort + "/inttest/shutdown")
	if err != nil {
		t.Fatalf("shutdown request failed: %v", err)
	}
	resp.Body.Close()

	// Wait for plain TCP forwarding to start
	deadline := time.Now().Add(3 * time.Second)
	var conn net.Conn
	for time.Now().Before(deadline) {
		conn, err = net.DialTimeout("tcp", "127.0.0.1:"+gitarPort, 500*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("plain TCP connection failed after TLS shutdown: %v", err)
	}
	defer conn.Close()

	// Verify echo through plain TCP
	msg := "hello after tls shutdown\n"
	conn.Write([]byte(msg))
	buf := make([]byte, len(msg))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(buf) != msg {
		t.Errorf("expected %q, got %q", msg, string(buf))
	}
}

func TestIntegrationNonTLSPortForwarding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)

	echoLn, echoPort := echoServer(t)
	defer echoLn.Close()

	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "inttest",
		"-p", gitarPort,
		"-f", echoPort,
		"--copy=false",
	)
	defer cleanup()

	httpClient := &http.Client{Timeout: 1 * time.Second}
	waitForHTTP(t, httpClient, "http://127.0.0.1:"+gitarPort+"/inttest/alias", 5*time.Second)

	resp, err := httpClient.Get("http://127.0.0.1:" + gitarPort + "/inttest/shutdown")
	if err != nil {
		t.Fatalf("shutdown request failed: %v", err)
	}
	resp.Body.Close()

	deadline := time.Now().Add(3 * time.Second)
	var conn net.Conn
	for time.Now().Before(deadline) {
		conn, err = net.DialTimeout("tcp", "127.0.0.1:"+gitarPort, 500*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("plain TCP connection failed: %v", err)
	}
	defer conn.Close()

	msg := "hello after http shutdown\n"
	conn.Write([]byte(msg))
	buf := make([]byte, len(msg))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(buf) != msg {
		t.Errorf("expected %q, got %q", msg, string(buf))
	}
}

// ──────────────────────────────────────────────
// File transfer: pull, push, pullr
// ──────────────────────────────────────────────

// createServedDir creates a directory tree to serve:
//
//	served/
//	  file1.txt
//	  subdir/
//	    file2.txt
//	    nested/
//	      file3.txt
func createServedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{
		"file1.txt",
		"subdir/file2.txt",
		"subdir/nested/file3.txt",
	} {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("content of "+rel+"\n"), 0644)
	}
	return dir
}

func TestIntegrationPullSingleFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)
	serveDir := createServedDir(t)
	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "test",
		"-p", gitarPort,
		"-d", serveDir,
		"--copy=false",
		"--completion=false",
	)
	defer cleanup()

	httpClient := &http.Client{Timeout: 2 * time.Second}
	baseURL := "http://127.0.0.1:" + gitarPort + "/test"
	waitForHTTP(t, httpClient, baseURL+"/alias", 5*time.Second)

	// Pull a single file
	resp, err := httpClient.Get(baseURL + "/pull/file1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "content of file1.txt\n" {
		t.Errorf("pull file1.txt: got %q", string(body))
	}

	// Pull nested file
	resp2, err := httpClient.Get(baseURL + "/pull/subdir/nested/file3.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	if string(body2) != "content of subdir/nested/file3.txt\n" {
		t.Errorf("pull nested file: got %q", string(body2))
	}
}

func TestIntegrationPushFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)
	serveDir := createServedDir(t)
	uploadDir := t.TempDir()
	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "test",
		"-p", gitarPort,
		"-d", serveDir,
		"-u", uploadDir,
		"--copy=false",
		"--completion=false",
	)
	defer cleanup()

	httpClient := &http.Client{Timeout: 2 * time.Second}
	baseURL := "http://127.0.0.1:" + gitarPort + "/test"
	waitForHTTP(t, httpClient, baseURL+"/alias", 5*time.Second)

	// Push a file via multipart POST
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "uploaded.txt")
	fw.Write([]byte("pushed content\n"))
	mw.Close()

	resp, err := httpClient.Post(baseURL+"/push", mw.FormDataContentType(), &body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Verify file landed in upload dir
	got, err := os.ReadFile(filepath.Join(uploadDir, "uploaded.txt"))
	if err != nil {
		t.Fatalf("pushed file not found: %v", err)
	}
	if string(got) != "pushed content\n" {
		t.Errorf("pushed file content: got %q", string(got))
	}
}

func TestIntegrationDirectoryListing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)
	serveDir := createServedDir(t)
	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "test",
		"-p", gitarPort,
		"-d", serveDir,
		"--copy=false",
		"--completion=false",
	)
	defer cleanup()

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // follow redirects (301 → listing)
		},
	}
	baseURL := "http://127.0.0.1:" + gitarPort + "/test"
	waitForHTTP(t, httpClient, baseURL+"/alias", 5*time.Second)

	// Request a directory — should get a listing containing links
	resp, err := httpClient.Get(baseURL + "/pull/subdir/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	listing := string(body)

	// The file server listing should contain links to the entries
	for _, want := range []string{"file2.txt", "nested/"} {
		if !strings.Contains(listing, want) {
			t.Errorf("directory listing missing %q; got:\n%s", want, listing)
		}
	}
}

// TestIntegrationPullrWithShell runs the actual pullr shell function against
// a live gitar server, using both bash and zsh (if available).
func TestIntegrationPullrWithShell(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)
	serveDir := createServedDir(t)
	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "test",
		"-p", gitarPort,
		"-d", serveDir,
		"--copy=false",
		"--completion=false",
	)
	defer cleanup()

	httpClient := &http.Client{Timeout: 2 * time.Second}
	baseURL := "http://127.0.0.1:" + gitarPort + "/test"
	waitForHTTP(t, httpClient, baseURL+"/alias", 5*time.Second)

	// Fetch the alias script
	resp, err := httpClient.Get(baseURL + "/alias")
	if err != nil {
		t.Fatal(err)
	}
	aliasScript, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	shells := []string{"bash"}
	if _, err := exec.LookPath("zsh"); err == nil {
		shells = append(shells, "zsh")
	}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			workDir := t.TempDir()

			// Build a script that sources the aliases and runs pullr
			// on the "subdir" directory, then verifies the files exist.
			script := string(aliasScript) + "\n" + fmt.Sprintf(`
cd "%s"
pullr subdir
# Verify files were downloaded
if [ ! -f "subdir/file2.txt" ]; then
  echo "FAIL: subdir/file2.txt missing"
  exit 1
fi
if [ ! -f "subdir/nested/file3.txt" ]; then
  echo "FAIL: subdir/nested/file3.txt missing"
  exit 1
fi
# Verify content
content=$(cat "subdir/file2.txt")
if [ "$content" != "content of subdir/file2.txt" ]; then
  echo "FAIL: wrong content in file2.txt: $content"
  exit 1
fi
content=$(cat "subdir/nested/file3.txt")
if [ "$content" != "content of subdir/nested/file3.txt" ]; then
  echo "FAIL: wrong content in file3.txt: $content"
  exit 1
fi
echo "OK"
`, workDir)

			cmd := exec.Command(shell)
			cmd.Stdin = strings.NewReader(script)
			cmd.Dir = workDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s pullr failed: %v\noutput:\n%s", shell, err, out)
			}
			if !strings.Contains(string(out), "OK") {
				t.Errorf("%s pullr unexpected output:\n%s", shell, out)
			}
		})
	}
}

func TestIntegrationAliasEndpointValid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := buildGitar(t)
	serveDir := createServedDir(t)
	gitarPort := freePort(t)

	_, cleanup := startGitar(t, binPath,
		"-e", "127.0.0.1",
		"-s", "test",
		"-p", gitarPort,
		"-d", serveDir,
		"--copy=false",
		"--completion=false",
	)
	defer cleanup()

	httpClient := &http.Client{Timeout: 2 * time.Second}
	baseURL := "http://127.0.0.1:" + gitarPort + "/test"
	waitForHTTP(t, httpClient, baseURL+"/alias", 5*time.Second)

	resp, err := httpClient.Get(baseURL + "/alias")
	if err != nil {
		t.Fatal(err)
	}
	script, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Verify script contains all expected functions
	for _, fn := range []string{"pull()", "pullr()", "push()", "pushr()", "status()", "getFiles()", "isDir()", "gtree()"} {
		if !strings.Contains(string(script), fn) {
			t.Errorf("alias script missing function: %s", fn)
		}
	}

	// Verify script contains the correct URL
	if !strings.Contains(string(script), "127.0.0.1:"+gitarPort+"/test") {
		t.Error("alias script does not contain the expected server URL")
	}
}
