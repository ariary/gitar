package gitar

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/go-utils/pkg/color"
)

//PortForwarding: forward all tcp port to specified port in config.
// Always uses plain TCP regardless of the --tls flag: TLS is only
// for the HTTP server phase. After shutdown the forwarder handles
// raw TCP so that reverse shells, netcat, etc. work directly.
func PortForwarding(config *config.Config) {
	fmt.Println()
	host := strings.Split(config.Url, "/")[2]
	fmt.Println(color.Info("Redirect all tcp traffic:"), host, "⏩ localhost:"+config.RedirectedPort)

	proxy, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := proxy.Accept()
		if err != nil {
			panic(err)
		}

		go handleRequest(conn, config.RedirectedPort)
	}
}

//handleRequest: forward connection to target service
func handleRequest(conn net.Conn, redirectedPort string) {
	targetService, err := net.Dial("tcp", "127.0.0.1:"+redirectedPort)
	if err != nil {
		panic(err)
	}
	fmt.Printf(color.Italic(color.Info("Forward connection from '%s' to '%s'\n")), conn.RemoteAddr().String(), targetService.RemoteAddr().String())
	go copyIO(conn, targetService)
	go copyIO(targetService, conn)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
