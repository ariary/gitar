package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/gitar/pkg/gitar"
	"github.com/ariary/gitar/pkg/send"
	"github.com/ariary/gitar/pkg/utils"
	"github.com/ariary/gitar/pkg/webhook"
	"github.com/ariary/go-utils/pkg/color"
	"github.com/spf13/cobra"
)

func main() {
	var detectExternal, windows, bidirectional, copyArg, tls, completion, noRun bool
	var serverIp, bidiDir, port, dlDir, upDir, certDir, aliasUrl, secret string

	//CMD ROOT
	var rootCmd = &cobra.Command{Use: "gitar",
		Short: "Launch an HTTP server to ease file sharing",
		Run: func(cmd *cobra.Command, args []string) {
			// Init
			config := gitar.InitGitar(serverIp, detectExternal, windows, bidirectional, bidiDir, port, dlDir, upDir, copyArg, tls, certDir, completion, aliasUrl, secret, noRun)

			//Set up messages
			//setUpMsgLinux := "curl -s " + url + "/alias > /tmp/alias && . /tmp/alias && rm /tmp/alias"
			gitar.SetUpMessage(config)

			// Launch
			gitar.LaunchGitar(config)
		},
	}

	// root flags
	rootCmd.Flags().StringVarP(&serverIp, "external", "e", "", "specify server external reachable ip/url")
	rootCmd.Flags().BoolVarP(&detectExternal, "detect-external", "i", false, "detect external ip and use it for gitar shortcut. If use with -e, the value of -e flag will be overwritten")
	rootCmd.Flags().BoolVarP(&windows, "windows", "w", false, "specify that the target machine is a window")
	rootCmd.Flags().StringVarP(&bidiDir, "bidi", "b", "", "bidirectional exchange: push file on target from the attacker machine without installing anything on target")
	rootCmd.Flags().StringVarP(&port, "port", "p", "9237", "specify HTTP server port")
	rootCmd.Flags().StringVarP(&dlDir, "dl-dir", "d", ".", "point to the directory of static files to serve")
	rootCmd.Flags().StringVarP(&upDir, "up-dir", "u", "./", "point to the directory where file are uploaded")
	rootCmd.Flags().BoolVarP(&copyArg, "copy", "c", true, "copy gitar set up command to clipboard (xclip required). True by default, disable with --copy=false")
	rootCmd.Flags().BoolVarP(&tls, "tls", "t", false, "use TLS (HTTPS server)")
	rootCmd.Flags().StringVarP(&certDir, "certs", "x", os.Getenv("HOME")+"/.gitar/certs", "point to the cert directory (use with --tls)")
	rootCmd.Flags().BoolVarP(&completion, "completion", "m", true, "enable completion for target machine (enabled by default). Works if target shell is bash, zsh")
	rootCmd.Flags().StringVarP(&aliasUrl, "alias-override-url ", "a", "", "override url in /alias endpoint (useful if gitar server is behind a proxy)")
	rootCmd.Flags().StringVarP(&secret, "secret", "s", "", "provide the secret that will prefix URL paths. (by default: auto-generated)")
	rootCmd.Flags().BoolVarP(&noRun, "dry-run", "", false, "do not launch gitar server, only return command to load shortcuts")

	//CMD SEND
	var host string
	var isDir, verbose, last bool
	var sendCmd = &cobra.Command{
		Use:                   "send",
		Short:                 "directly send a file to a target using different options",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		// Omitting the Run (and RunE) field from the cobra.Command will make it a requirement for a valid subcommand to be given
	}

	// send flags
	sendCmd.PersistentFlags().BoolVarP(&isDir, "dir", "d", false, "send a directory")
	sendCmd.PersistentFlags().BoolVarP(&last, "last", "l", false, "apply previous settings to send the file (no user interation)")
	sendCmd.PersistentFlags().BoolVarP(&verbose, "show", "v", false, "only show the command do not execute it")
	sendCmd.PersistentFlags().StringVarP(&host, "host", "t", "", "specify host/target url to send the file")
	sendCmd.PersistentFlags().StringVarP(&port, "port", "p", "", "specify target port")

	// SUBCMD SCP
	//TODO target directorie
	var user, password, keyfile string
	var withKey bool
	var scpCmd = &cobra.Command{
		Use:     "scp",
		Aliases: []string{"ssh"},
		Short:   "send with scp",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]

			//special case: directory
			var flags string
			fileInfo, _ := os.Stat(file)
			if fileInfo.IsDir() {
				utils.Tar(file, filepath.Base(file)+".tar.gz")
				file = filepath.Base(file) + ".tar.gz"
				defer func() {
					err := os.Remove(file)
					if err != nil {
						fmt.Printf("failed to delete archive %s: %s\n", file, err)
					}
				}()

				// scp -r, wait for go-scp library to handle it see https://github.com/bramvdbogaerde/go-scp/issues/61
				//when it's done, comment above code
				flags = " -r "
			}
			cfg := &config.ConfigScp{}
			send.ReadLastScpConfig(cfg)
			if !last {
				//TO DO: determine which flags are already provided
				// Send them to Asku user Input to determine if input is necessary
				send.AskUserInputForScp(cfg, *cmd.Flags())
			}

			//scp -P 2222 go.mod root@192.168.1.100:/tmp/
			//remoteTarget := "/home/" + cfg.User
			remoteTarget := "/tmp/" + filepath.Base(file)
			command := "scp -P " + cfg.Port + " " + flags + file + " " + cfg.User + "@" + cfg.Host + ":" + remoteTarget
			fmt.Println(color.Dim(command))
			if verbose {
				return
			}
			send.ExecScp(cfg, file, remoteTarget)

			send.UpdateScpConfig(cfg)
		},
	}
	//scp flags
	scpCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "specify ssh user")
	scpCmd.PersistentFlags().StringVarP(&password, "password", "x", "", "specify ssh user password")
	scpCmd.PersistentFlags().StringVarP(&keyfile, "key", "i", "", "specify ssh private key file")
	scpCmd.PersistentFlags().BoolVarP(&withKey, "with-key", "k", false, "specify if authentatication scheme udes key instead of password")

	//CMD WEBHOOK
	var proxy string
	var webhookCmd = &cobra.Command{
		Use:   "webhook",
		Short: "HTTP handler to observe incoming request",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			history := webhook.History{}
			if proxy == "" {
				//use middleware
				mux := http.NewServeMux()

				finalHandler := http.HandlerFunc(webhook.FinalHandler)
				mux.Handle("/", webhook.Middleware(finalHandler, &history))

				fmt.Println("HTTP webhook  listening on", port, "...")
				err := http.ListenAndServe(":"+port, mux)
				log.Fatal(err)
			} else {
				// as a reverse proxy see https://blog.joshsoftware.com/2021/05/25/simple-and-powerful-reverseproxy-in-go/
				fmt.Println(color.WhiteForeground("🔄"), color.Italic("Reverse proxy mode"))
				proxy, err := webhook.NewProxy(proxy, &history)
				if err != nil {
					panic(err)
				}
				fmt.Println("HTTP webhook listening on", port, "...")
				http.HandleFunc("/", webhook.ProxyRequestHandler(proxy))
				log.Fatal(http.ListenAndServe(":"+port, nil))
			}

		},
	}

	//webhook flags
	webhookCmd.PersistentFlags().StringVarP(&proxy, "proxy", "", "", "use webhook as a reverse proxy")
	webhookCmd.PersistentFlags().StringVarP(&port, "port", "p", "9292", "specify webhook port")
	//response overriding (status code and headers)
	//request filter (only show if), method,header,param
	//request logs: header,param
	//show full request
	//serve file also

	// SUBCOMMANDS
	sendCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(webhookCmd)
	rootCmd.Execute()

}
