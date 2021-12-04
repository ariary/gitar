package main

import (
	"flag"
	"fmt"
	"gitar/pkg/handlers"
	"gitar/pkg/utils"
	"log"
	"net/http"
)

func main() {
	serverIp := flag.String("e", "127.0.0.1", "Server external reachable ip")
	port := flag.String("p", "9237", "Port to serve on")
	directory := flag.String("d", ".", "Point to the directory of static file to host")
	copyArg := flag.Bool("copy", true, "Copy gitar set up command to clipboard (xclip required)")
	tls := flag.Bool("tls", false, "Use HTTPS server (TLS)")
	flag.Parse()

	handlers.InitHandlers(*directory, *serverIp, *port)

	//Set up messages
	setUpMsg := "curl -s http://" + *serverIp + ":" + *port + "/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias"
	fmt.Println("Launch it on remote to set up gitar exchange:")
	fmt.Println(setUpMsg)
	if *copyArg {
		utils.Check(utils.Copy(setUpMsg), "")
	}

	//Listen
	var err error
	if *tls {
		fmt.Println("toto")
		err = http.ListenAndServeTLS(":"+*port, "server.crt", "server.key", nil)
	} else {
		err = http.ListenAndServe(":"+*port, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
