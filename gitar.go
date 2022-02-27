package main

import (
	"flag"
	"fmt"
	"gitar/pkg/config"
	"gitar/pkg/handlers"
	"log"
	"net/http"
	"os"

	"github.com/ariary/go-utils/pkg/check"
	"github.com/ariary/go-utils/pkg/clipboard"
)

func main() {
	serverIp := flag.String("e", "127.0.0.1", "Server external reachable ip")
	port := flag.String("p", "9237", "Port to serve on")
	dlDir := flag.String("d", ".", "Point to the directory of static file to serve")
	upDir := flag.String("u", "./", "Point to the directory where file are uploaded")
	copyArg := flag.Bool("copy", true, "Copy gitar set up command to clipboard (xclip required)")
	tls := flag.Bool("tls", false, "Use HTTPS server (TLS)")
	certDir := flag.String("c", os.Getenv("HOME")+"/.gitar/certs", "Point to the cert directory")
	completion := flag.Bool("completion", true, "Enable completion for target machine") //False for /bin/sh (don't have complete)
	aliasUrl := flag.String("alias-override-url", "", "Override url in /alias endpoint (useful if gitar server is behind a proxy)")
	noRun := flag.Bool("dry-run", false, "Do not launch gitar server, only return command to load shortcuts")

	flag.Parse()

	cfg := &config.Config{ServerIP: *serverIp, Port: *port, DownloadDir: *dlDir, UploadDir: *upDir + "/", IsCopied: *copyArg, Tls: *tls, AliasUrl: *aliasUrl, Completion: *completion}

	//Set up messages
	ip := cfg.ServerIP
	p := cfg.Port
	var protocol string
	if cfg.Tls {
		protocol = "-k https://"
	} else {
		protocol = "http://"
	}
	url := protocol + ip + ":" + p
	if cfg.AliasUrl != "" {
		url = cfg.AliasUrl
	}

	//setUpMsg := "curl -s " + url + "/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias"
	setUpMsg := "curl -s " + url + "/alias > /tmp/alias && . /tmp/alias && rm /tmp/alias"
	if !*noRun {
		fmt.Println("Launch it on remote to set up gitar exchange:")
	}
	fmt.Println(setUpMsg)
	if *copyArg {
		check.Check(clipboard.Copy(setUpMsg), "")
	}

	if *noRun {
		os.Exit(0)
	}

	//handlers
	handlers.InitHandlers(cfg)

	//Listen
	var err error
	if cfg.Tls {
		err = http.ListenAndServeTLS(":"+cfg.Port, *certDir+"/server.crt", *certDir+"/server.key", nil)
	} else {
		err = http.ListenAndServe(":"+cfg.Port, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
