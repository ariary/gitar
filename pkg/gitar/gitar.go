package gitar

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/go-utils/pkg/clipboard"
	"github.com/ariary/go-utils/pkg/color"
	"github.com/ariary/go-utils/pkg/host"
)

//InitGitar: initialize configuration structs according to flags
func InitGitar(serverIp string, detectExternal bool, windows bool, bidirectional bool, bidiDir string, port string, dlDir string, upDir string, copyArg bool, tls bool, certDir string, completion bool, aliasUrl string, secret string, noRun bool) (cfg *config.Config) {
	if serverIp == "" { //no ip provided
		var err error
		if detectExternal { //use external IP
			serverIp, err = host.GetExternalIP()
			if err != nil {
				fmt.Println("Failed to detect external ip (dig):", err)
				os.Exit(1)
			}
		} else { //use hostname ip
			serverIp, err = host.GetHostIP()
			if err != nil {
				fmt.Println("Failed to detect host ip (hostname):", err)
				os.Exit(1)
			}
		}

	} else if detectExternal {
		var err error
		serverIp, err = host.GetExternalIP()
		if err != nil {
			fmt.Println("Failed to detect external ip (dig):", err)
			os.Exit(1)
		}

	}

	//Secret generation
	if secret == "" {
		//generate random string
		//*secret = encryption.GenerateRandom()
		rand.Seed(time.Now().UnixNano())
		var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789-_=")
		b := make([]rune, 7)
		for i := range b {
			b[i] = characters[rand.Intn(len(characters))]
		}
		secret = string(b)
	}

	//bidirectional
	var mktempDir string
	if bidiDir == "" {
		bidiDir = "/tmp"
	}
	if bidirectional {
		mktemp, err := exec.Command("mktemp", "-p", bidiDir, "-d").Output()
		if err != nil {
			fmt.Println(err)
		} else {
			//Configure bidi directory
			mktempDir = strings.ReplaceAll(string(mktemp), "\n", "")

			//Configure alias for host
			//alias, err := exec.Command("mktemp", "--suffix=gitar").Output()
			alias, err := exec.Command("mktemp", "-p", bidiDir, "gitarXXXXXX").Output() //alpine compliant
			if err != nil {
				fmt.Println(err)
			} else {
				aliasFile := strings.ReplaceAll(string(alias), "\n", "")
				hostAliases := `
		push(){
			cp $1 ` + mktempDir + ` 
		}
		`
				err = os.WriteFile(string(aliasFile), []byte(hostAliases), 0644)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
	}

	//Url construction
	ip := serverIp
	p := port
	var protocol string
	if tls {
		protocol = "-k https://"
	} else {
		protocol = "http://"
	}

	var url string
	if aliasUrl != "" {
		url = aliasUrl + "/" + secret
	} else {
		url = protocol + ip + ":" + p + "/" + secret
	}

	cfg = &config.Config{ServerIP: serverIp, Port: port, DownloadDir: dlDir, UploadDir: upDir + "/", IsCopied: copyArg, Tls: tls, Url: url, Completion: completion, Secret: secret, BidirectionalDir: mktempDir, Windows: windows, CertDir: certDir, NoRun: noRun}

	return cfg
}

//SetUpMessages: set up message that will be output to help in the final set up of gitar on target
func SetUpMessage(config *config.Config) {
	setUpMsgLinux := "source <(curl -s " + config.Url + "/alias)"
	setUpMsg := setUpMsgLinux
	if config.Windows {
		setUpMsgWindows := color.BlueForeground("Powershell:")
		setUpMsgWindows += "\n(Invoke-WebRequest " + config.Url + "/aliaswinpsinvokeweb).Content | iex "
		setUpMsgWindows += color.Dim("\nInvoke-WebRequest " + config.Url + "/aliaswinpsinvokeweb -OutFile ./alias.ps1 && . ./alias.ps1 && del ./alias.ps1")
		// setUpMsgWindows += "\nInvoke-RestMethod " + config.Url + "/aliaswinpsinvokeres | iex " + color.Dim("(in-memory execution)")
		// setUpMsgWindows += "\nInvoke-RestMethod " + config.Url + "/aliaswinpsinvokeres > ./alias.ps1 && . ./alias.ps1 && del ./alias.ps1 "
		setUpMsgWindows += "\n"
		setUpMsgWindows += color.YellowForeground("CMD.exe:")
		setUpMsgWindows += "\nto do.."
		setUpMsg = setUpMsgWindows
	}

	if !config.NoRun {
		fmt.Println("Set up gitar exchange on remote:")
	}
	fmt.Println(setUpMsg)
	if config.NoRun {
		os.Exit(0)
	}

	if config.IsCopied {
		clipboard.Copy(setUpMsg)
	}
}

func LaunchGitar(config *config.Config) {
	//handlers
	InitHandlers(config)

	//Listen
	var err error
	if config.Tls {
		err = http.ListenAndServeTLS(":"+config.Port, config.CertDir+"/server.crt", config.CertDir+"/server.key", nil)
	} else {
		err = http.ListenAndServe(":"+config.Port, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
