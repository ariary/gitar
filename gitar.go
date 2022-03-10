package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/gitar/pkg/handlers"

	"github.com/ariary/go-utils/pkg/clipboard"
)

func main() {
	var detectExternal bool

	serverIp := flag.String("e", "", "Server external reachable ip")
	flag.BoolVar(&detectExternal, "ext", false, "detect external ip and use it for gitar shortcut. If use with -e, the value of -e flag will be overwritten")
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

	// external IP checks
	if *serverIp == "" { //no ip provided
		var err error
		if detectExternal { //use external IP
			*serverIp, err = getExternalIP()
			if err != nil {
				fmt.Println("Failed to detect external ip (dig):", err)
				os.Exit(1)
			}
		} else { //use hostname ip
			*serverIp, err = getHostIP()
			if err != nil {
				fmt.Println("Failed to detect host ip (hostname):", err)
				os.Exit(1)
			}
		}

	} else if detectExternal {
		var err error
		*serverIp, err = getExternalIP()
		if err != nil {
			fmt.Println("Failed to detect external ip (dig):", err)
			os.Exit(1)
		}

	}

	//Url construction
	ip := *serverIp
	p := *port
	var protocol string
	if *tls {
		protocol = "-k https://"
	} else {
		protocol = "http://"
	}

	var url string
	if *aliasUrl != "" {
		url = *aliasUrl
	} else {
		url = protocol + ip + ":" + p
	}

	cfg := &config.Config{ServerIP: *serverIp, Port: *port, DownloadDir: *dlDir, UploadDir: *upDir + "/", IsCopied: *copyArg, Tls: *tls, Url: url, Completion: *completion}

	//Set up messages
	//setUpMsg := "curl -s " + url + "/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias"
	setUpMsg := "curl -s " + url + "/alias > /tmp/alias && . /tmp/alias && rm /tmp/alias"
	if !*noRun {
		fmt.Println("Launch it on remote to set up gitar exchange:")
	}
	fmt.Println(setUpMsg)
	if *noRun {
		os.Exit(0)
	}

	if *copyArg {
		clipboard.Copy(setUpMsg)
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

func getExternalIP() (ip string, err error) {
	cmd := exec.Command("dig", "@resolver4.opendns.com", "myip.opendns.com", "+short")
	ipB, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ip = string(ipB)
	ip = strings.ReplaceAll(ip, "\n", "")
	return ip, err
}

func getHostIP() (ip string, err error) {
	cmd := exec.Command("hostname", "-I")
	ipB, err := cmd.Output()
	if err != nil {
		//retry with -i
		cmd := exec.Command("hostname", "-i")
		ipB, err := cmd.Output()
		if err != nil {
			return "", err
		}
		ip = strings.ReplaceAll(string(ipB), "\n", "")
		return ip, nil
	}
	//Only take first result
	//r := bytes.NewReader(ipB)
	//reader := bufio.NewReader(r)
	//line, _, err := reader.ReadLine()
	//ip = string(line)
	//ip = strings.ReplaceAll(ip, " ", "")
	ip = strings.Fields(string(ipB))[0]
	return ip, err
}
