package gitar

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// echoServer accepts connections and echoes data back. Returns the listener
// and a cleanup function.
func echoServer(t *testing.T) (net.Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()
	return ln, port
}

// forwardingProxy creates a TCP listener that forwards each connection
// through handleRequest to the given target port.
func forwardingProxy(t *testing.T, targetPort string) (net.Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleRequest(conn, targetPort)
		}
	}()
	return ln, port
}

func TestHandleRequestForwardsPlainTCP(t *testing.T) {
	target, targetPort := echoServer(t)
	defer target.Close()

	proxy, proxyPort := forwardingProxy(t, targetPort)
	defer proxy.Close()

	client, err := net.DialTimeout("tcp", "127.0.0.1:"+proxyPort, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := "hello from plain tcp\n"
	if _, err := client.Write([]byte(msg)); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(msg))
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(buf) != msg {
		t.Errorf("expected %q, got %q", msg, string(buf))
	}
}

func TestHandleRequestBidirectional(t *testing.T) {
	// Target server: reads a line, then sends its own message, then echoes
	target, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	_, targetPort, _ := net.SplitHostPort(target.Addr().String())

	serverReply := "reply from server\n"
	go func() {
		conn, err := target.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Read client message
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		// Send server reply + echo back the client message
		conn.Write([]byte(serverReply))
		conn.Write(buf[:n])
	}()

	proxy, proxyPort := forwardingProxy(t, targetPort)
	defer proxy.Close()

	client, err := net.DialTimeout("tcp", "127.0.0.1:"+proxyPort, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	clientMsg := "hello from client\n"
	client.Write([]byte(clientMsg))

	// Read both server reply and echoed message
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	all, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	got := string(all)
	want := serverReply + clientMsg
	if got != want {
		t.Errorf("bidirectional data mismatch:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestHandleRequestConcurrentConnections(t *testing.T) {
	target, targetPort := echoServer(t)
	defer target.Close()

	proxy, proxyPort := forwardingProxy(t, targetPort)
	defer proxy.Close()

	const numClients = 10
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", "127.0.0.1:"+proxyPort, 2*time.Second)
			if err != nil {
				t.Errorf("client %d: dial failed: %v", id, err)
				return
			}
			defer conn.Close()

			msg := []byte("ping from client\n")
			if _, err := conn.Write(msg); err != nil {
				t.Errorf("client %d: write failed: %v", id, err)
				return
			}

			buf := make([]byte, len(msg))
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Errorf("client %d: read failed: %v", id, err)
				return
			}
			if string(buf) != string(msg) {
				t.Errorf("client %d: expected %q, got %q", id, msg, buf)
			}
		}(i)
	}
	wg.Wait()
}

func TestHandleRequestLargePayload(t *testing.T) {
	target, targetPort := echoServer(t)
	defer target.Close()

	proxy, proxyPort := forwardingProxy(t, targetPort)
	defer proxy.Close()

	client, err := net.DialTimeout("tcp", "127.0.0.1:"+proxyPort, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send 1MB of data
	payload := make([]byte, 1<<20)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	go func() {
		client.Write(payload)
	}()

	buf := make([]byte, len(payload))
	client.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("read large payload: %v", err)
	}
	for i := range payload {
		if buf[i] != payload[i] {
			t.Fatalf("mismatch at byte %d: got %d, want %d", i, buf[i], payload[i])
		}
	}
}
