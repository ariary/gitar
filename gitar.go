package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/gitar/pkg/handlers"

	"github.com/ariary/go-utils/pkg/clipboard"
)

var usage = `Usage of gitar: gitar [flags]
Launch an HTTP server to ease file sharing
  -e                       external reachable ip to server the HTTP server
  -p                       specify HTTP server port
  --ext                    detect the external IP to use
  -d                       point to the directory of static file to serve
  -u                       point to the directory where file are uploaded
  --copy                   copy gitar set up command to clipboard (xclip required). True by default, disable with --copy=false
  --tls                    use TLS (HTTPS server)
  -c                       point to the cert directory (use with --tls)
  --completion             enable completion for target machine (enabled by default). Works if target shell is bash, zsh
  --alias-override-url     override url in /alias endpoint (useful if gitar server is behind a proxy)
  --secret                 provide the secret that will prefix URL paths. (by default: auto-generated)
  --dry-run                do not launch gitar server, only return command to load shortcuts
  --windows		           specify that the target machine is a windows

  -h, --help                  prints help information 
`

func main() {
	var detectExternal, windows bool

	serverIp := flag.String("e", "", "Server external reachable ip")
	flag.BoolVar(&detectExternal, "ext", false, "Detect external ip and use it for gitar shortcut. If use with -e, the value of -e flag will be overwritten")
	flag.BoolVar(&windows, "windows", false, "Target machine is a windows (copy paste windows shortcuts)")
	port := flag.String("p", "9237", "Port to serve on")
	dlDir := flag.String("d", ".", "Point to the directory of static file to serve")
	upDir := flag.String("u", "./", "Point to the directory where file are uploaded")
	copyArg := flag.Bool("copy", true, "Copy gitar set up command to clipboard (xclip required)")
	tls := flag.Bool("tls", false, "Use HTTPS server (TLS)")
	certDir := flag.String("c", os.Getenv("HOME")+"/.gitar/certs", "Point to the cert directory")
	completion := flag.Bool("completion", true, "Enable completion for target machine") //False for /bin/sh (don't have complete)
	aliasUrl := flag.String("alias-override-url", "", "Override url in /alias endpoint (useful if gitar server is behind a proxy)")
	secret := flag.String("secret", "", "Provide a secret that will prefix URL paths. (by default: auto-generated)")
	noRun := flag.Bool("dry-run", false, "Do not launch gitar server, only return command to load shortcuts")

	flag.Usage = func() { fmt.Print(usage) }
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

	//Sercet generation
	if *secret == "" {
		//generate random string
		//*secret = encryption.GenerateRandom()
		rand.Seed(time.Now().UnixNano())
		var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789-_=")
		b := make([]rune, 7)
		for i := range b {
			b[i] = characters[rand.Intn(len(characters))]
		}
		*secret = string(b)
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
		url = *aliasUrl + "/" + *secret
	} else {
		url = protocol + ip + ":" + p + "/" + *secret
	}

	cfg := &config.Config{ServerIP: *serverIp, Port: *port, DownloadDir: *dlDir, UploadDir: *upDir + "/", IsCopied: *copyArg, Tls: *tls, Url: url, Completion: *completion, Secret: *secret}

	//Set up messages
	//setUpMsgLinux := "curl -s " + url + "/alias > /tmp/alias && . /tmp/alias && rm /tmp/alias"
	setUpMsgLinux := "source <(curl -s " + url + "/alias)"
	setUpMsgWindows := "curl -s " + url + "/aliaswin > ./alias && doskey /macrofile=alias && del alias"
	setUpMsg := setUpMsgLinux
	if windows {
		setUpMsg = setUpMsgWindows
	}
	fmt.Println(cfg.Secret)
	if !*noRun {
		fmt.Println("Set up gitar exchange on remote:")
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
		ip = string(ipB)
		return ip, nil
	}

	//Only take first result
	r := bytes.NewReader(ipB)
	reader := bufio.NewReader(r)
	line, _, err := reader.ReadLine()
	ip = string(line)
	ip = strings.ReplaceAll(ip, " ", "")
	return ip, err
}
