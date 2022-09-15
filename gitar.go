package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	var serverIp, bidiDir, port, dlDir, upDir, certDir, aliasUrl, secret, redirectedPort string

	//CMD ROOT
	var rootCmd = &cobra.Command{Use: "gitar",
		Short: "Launch an HTTP server to ease file sharing",
		Run: func(cmd *cobra.Command, args []string) {
			// Init
			config := gitar.InitGitar(serverIp, detectExternal, windows, bidirectional, bidiDir, port, dlDir, upDir, copyArg, tls, certDir, completion, aliasUrl, secret, noRun, redirectedPort)

			//Set up messages
			//setUpMsgLinux := "curl -s " + url + "/alias > /tmp/alias && . /tmp/alias && rm /tmp/alias"
			gitar.SetUpMessage(config)

			// Launch gitar HTTP server
			gitar.LaunchGitar(config)

			// if port redirection is used, perform it
			gitar.PortForwarding(config)
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
	rootCmd.Flags().StringVarP(&redirectedPort, "port-forward", "f", "", "set-up handler to shutdown server, once server is shutdown all tcp traffic is redirect to specified port")

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
	var proxy, prefixStatic string
	var headers, statics []string
	cfg := config.ConfigWebHook{}
	var webhookCmd = &cobra.Command{
		Use:   "webhook",
		Short: "HTTP handler to observe incoming request",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.OverrideHeaders, cfg.DelHeaders = config.ParseHeadersFromCLI(headers)
			if proxy == "" {
				//use middleware
				mux := http.NewServeMux()

				if len(statics) == 0 {
					finalHandler := http.HandlerFunc(webhook.FinalProcessResponseHandler(&cfg))
					mux.Handle("/", webhook.Middleware(finalHandler, &cfg))
				} else {
					for i := 0; i < len(statics); i++ {
						fileHandler := http.StripPrefix("/"+prefixStatic, http.FileServer(http.Dir(statics[i])))
						finalHandler := webhook.ProcessResponseHandler(fileHandler, &cfg)
						mux.Handle("/", webhook.Middleware(finalHandler, &cfg))
					}
				}

				webhookBanner(cfg, port, statics, prefixStatic)
				err := http.ListenAndServe(":"+port, mux)
				log.Fatal(err)
			} else {
				// as a reverse proxy see https://blog.joshsoftware.com/2021/05/25/simple-and-powerful-reverseproxy-in-go/
				fmt.Println(color.WhiteForeground("🔄"), color.Italic("Reverse proxy mode (to "+proxy+")"))
				proxy, err := webhook.NewProxy(proxy, &cfg)
				if err != nil {
					panic(err)
				}
				if len(statics) > 0 {
					fmt.Println("--serve/-f option cannot be used with --proxy")
					os.Exit(92)
				}
				webhookBanner(cfg, port, statics, prefixStatic)
				http.HandleFunc("/", webhook.ProxyRequestHandler(proxy))
				log.Fatal(http.ListenAndServe(":"+port, nil))
			}

		},
	}

	//webhook flags
	webhookCmd.PersistentFlags().StringVarP(&proxy, "proxy", "", "", "use webhook as a reverse proxy")
	webhookCmd.PersistentFlags().StringVarP(&port, "port", "p", "9292", "specify webhook port")
	webhookCmd.PersistentFlags().StringSliceVarP(&cfg.Params, "params", "P", cfg.Params, "filter incoming request parameter. Can be used multiple times.")
	webhookCmd.PersistentFlags().BoolVarP(&cfg.FullBody, "body", "b", false, "print full body of POST request")
	webhookCmd.PersistentFlags().BoolVarP(&cfg.FullHeaders, "show-headers", "S", false, "print all the headers of the incoming request")
	webhookCmd.PersistentFlags().StringSliceVarP(&cfg.ReqHeaders, "request-header", "C", cfg.ReqHeaders, "catch request header Can be used multiple times")
	webhookCmd.PersistentFlags().StringSliceVarP(&headers, "header", "H", headers, "add/override response header (in form of name:value to add header OR to remove header: name:). Can be used multiple times.")
	webhookCmd.PersistentFlags().StringSliceVarP(&statics, "serve", "f", statics, "specifiy folder to serve static file. Can be used multiple times. (can't be used with proxy mode)")
	webhookCmd.PersistentFlags().StringVarP(&prefixStatic, "override-prefix", "o", prefixStatic, "specify the prefix path for static file. (if --serve is used)")
	//TODO: full request + status code

	// SUBCOMMANDS
	sendCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(webhookCmd)
	rootCmd.Execute()

}

func webhookBanner(cfg config.ConfigWebHook, port string, statics []string, prefix string) {
	//params
	if len(cfg.Params) > 0 {
		fmt.Println(color.BlueForeground("👁️ Catch request parameters:"))
		for i := 0; i < len(cfg.Params); i++ {
			fmt.Println("  • " + cfg.Params[i])
		}
		fmt.Println()
	}
	//header
	if len(cfg.Params) > 0 {
		fmt.Println(color.BlueForeground("👁️ Catch request headers:"))
		for i := 0; i < len(cfg.ReqHeaders); i++ {
			fmt.Println("  • " + cfg.ReqHeaders[i])
		}
		fmt.Println()
	}
	if len(cfg.DelHeaders) > 0 {
		fmt.Println(color.TealForeground("🗑️ Delete response headers:"))
		for i := 0; i < len(cfg.DelHeaders); i++ {
			fmt.Println("  • " + cfg.DelHeaders[i])
		}
		fmt.Println()
	}

	if len(cfg.OverrideHeaders) > 0 {
		fmt.Println(color.TealForeground("✍️ Override/add response headers:"))
		for header, value := range cfg.OverrideHeaders {
			fmt.Println("  • " + header + ": " + strings.Join(value, ","))
		}
		fmt.Println()
	}
	//static files
	if len(statics) > 0 {
		if prefix != "" {
			prefix = color.Italic(" (URL prefix path of static files: " + prefix + ")")
		}
		fmt.Println(color.YellowForeground("📁 Serving static folders:", prefix))
		for i := 0; i < len(statics); i++ {
			if path, err := filepath.Abs(statics[i]); err == nil {
				fmt.Println("  • " + path)
			}

		}
		fmt.Println()
	}
	fmt.Println("HTTP webhook  listening on", "0.0.0.0:"+port, "...")
}
