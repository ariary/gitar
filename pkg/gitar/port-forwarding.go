package gitar

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/go-utils/pkg/color"
)

//PortForwarding: forward all tcp port to specified port in config
func PortForwarding(config *config.Config) {
	fmt.Println()
	fmt.Println("Redirect all tcp traffic from", config.Port, "to", config.RedirectedPort)
	// signals := make(chan os.Signal, 1)
	// stop := make(chan bool)
	// signal.Notify(signals, os.Interrupt)
	// go func() {
	// 	for _ = range signals {
	// 		fmt.Println("\nReceived an interrupt, stopping forwarding...")
	// 		stop <- true
	// 	}
	// }()

	// // Incoming request (set up listener)
	// var incoming net.Listener
	// var err error
	// if config.Tls {
	// 	//TODO: tls traffic is forwarded but does not seem to be decrypted when forwad
	// 	cert, err := tls.LoadX509KeyPair(config.CertDir+"/server.crt", config.CertDir+"/server.key")
	// 	if err != nil {
	// 		log.Fatalf("server: loadkeys: %s", err)
	// 	}
	// 	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	// 	tlsConfig.Rand = rand.Reader
	// 	incoming, err = tls.Listen("tcp", ":"+config.Port, &tlsConfig)
	// 	if err != nil {
	// 		log.Fatalf("could not start (TLS) port-forwarding server  on %s: %v", config.Port, err)
	// 	}
	// } else {
	// 	incoming, err = net.Listen("tcp", ":"+config.Port)
	// 	if err != nil {
	// 		log.Fatalf("could not start port-forwarding server on %s: %v", config.Port, err)
	// 	}
	// }

	// client, err := incoming.Accept()
	// if err != nil {
	// 	log.Fatal("could not accept client connection", err)
	// }
	// defer client.Close()
	// //fmt.Printf(color.Italic(color.Info("Forward connection from '%v'!\n")), client.RemoteAddr())

	// targetService, err := net.Dial("tcp", "localhost:"+config.RedirectedPort)
	// if err != nil {
	// 	log.Fatal("could not connect to target service", err)
	// }
	// defer targetService.Close()
	// fmt.Printf(color.Italic(color.Info("Forward connection to '%v'!\n")), targetService.RemoteAddr())

	// go func() { io.Copy(targetService, client) }()
	// go func() { io.Copy(client, targetService) }()

	// <-stop

	// Incoming request (set up listener)
	var proxy net.Listener
	var err error
	if config.Tls {
		//TODO: tls traffic is forwarded but does not seem to be decrypted when forwad
		cert, err := tls.LoadX509KeyPair(config.CertDir+"/server.crt", config.CertDir+"/server.key")
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.Rand = rand.Reader
		proxy, err = tls.Listen("tcp", ":"+config.Port, &tlsConfig)
		if err != nil {
			log.Fatalf("could not start (TLS) port-forwarding server  on %s: %v", config.Port, err)
		}
	} else {
		proxy, err = net.Listen("tcp", ":"+config.Port)
		if err != nil {
			panic(err)
		}
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
