package gitar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	scp "github.com/bramvdbogaerde/go-scp"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/go-utils/pkg/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//ReadLastScpConfig: read last scp config to provide suggestions. (~/.gitar/scp_conf)
func ReadLastScpConfig(cfg *config.ConfigScp) {
	file, _ := ioutil.ReadFile(os.ExpandEnv("$HOME") + "/.gitar/scp_conf")

	_ = json.Unmarshal([]byte(file), &cfg)
}

//UpdateScpConfig: update config file with config. (~/.gitar/scp_conf)
func UpdateScpConfig(cfg *config.ConfigScp) {
	if file, err := json.MarshalIndent(cfg, "", " "); err != nil {
		fmt.Println("Error while updating scp configuration:", err)
	} else {
		if err = ioutil.WriteFile(os.ExpandEnv("$HOME")+"/.gitar/scp_conf", file, 0644); err != nil {
			fmt.Println("Error while updating scp configuration:", err)
		}
	}

}

//Ask user to provide information needed
func AskUserInputForScp(cfg *config.ConfigScp) {
	// host
	waitHostInput(cfg)

	// port
	var port, msg string
	var portInput string
	if cfg.Port == "" {
		port = "[22]"
	} else {
		port = cfg.Port
	}
	msg = color.Blue("»") + " Port:" + color.Cyan(port) + " "
	fmt.Printf(msg)
	fmt.Scanln(&portInput)
	if portInput == "" {
		if cfg.Port == "" {
			cfg.Port = "22"
		}
		//else nothing
	} else {
		cfg.Port = portInput
	}

	// username
	waitUsernameInput(cfg)

	// password
	waitPasswordInput(cfg)

}

func ExecScp(cfg *config.ConfigScp, localFilename string, remoteFilename string) {
	// initial try with https://github.com/povsister but some bug with tranfer with password
	clientConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client := scp.NewClient(cfg.Host+":"+cfg.Port, clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}

	f, err := os.Open(localFilename)
	if err != nil {
		fmt.Println("Error while opening file", localFilename, ":", err)
		os.Exit(92)
	}
	info, err := f.Stat()
	var mode string
	if err != nil {
		mode = "0655"
	} else {
		mode = "0" + strconv.FormatInt(int64(info.Mode().Perm()), 8)
	}
	// Close client connection after the file has been copied
	defer client.Close()
	// Close the file after it has been copied
	defer f.Close()

	err = client.CopyFromFile(context.Background(), *f, remoteFilename, mode)

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}

}

func waitHostInput(cfg *config.ConfigScp) {
	var hostInput string
	msg := color.Blue("»") + " Host:"
	if cfg.User != "" {
		msg += color.Cyan(cfg.User)
	}
	msg += " "
	fmt.Printf(msg)
	fmt.Scanln(&hostInput)
	if hostInput == "" {
		if cfg.Host == "" {
			waitHostInput(cfg)
		} else {
			return
		}
	} else {
		cfg.Host = hostInput
	}
}

func waitUsernameInput(cfg *config.ConfigScp) {
	var userInput string
	msg := color.Blue("»") + " User:"
	if cfg.User != "" {
		msg += color.Cyan(cfg.User)
	}
	msg += " "
	fmt.Printf(msg)
	fmt.Scanln(&userInput)
	if userInput == "" {
		if cfg.User == "" {
			waitUsernameInput(cfg)
		} else {
			return
		}
	} else {
		cfg.User = userInput
	}
}

func waitPasswordInput(cfg *config.ConfigScp) {

	msg := color.Blue("»") + " Password: "

	fmt.Printf(msg)
	password, err := terminal.ReadPassword(0)
	fmt.Println()
	if err != nil {
		fmt.Println("erro while typing password:", password)
	}
	if string(password) == "" {
		waitPasswordInput(cfg)
	} else {
		cfg.Password = string(password)
	}

}
