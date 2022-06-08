package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/gitar/pkg/gitar"
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

	// CMD SCP
	//TODO target directorie
	var user, password, keyfile string
	var scpCmd = &cobra.Command{
		Use:     "scp",
		Aliases: []string{"ssh"},
		Short:   "send with scp",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]
			var flags string
			if isDir {
				// utils.Tar(file, file+".tar.gz")
				// file = file + ".tar.gz"
				// defer func() {
				// 	err := os.Remove(file)
				// 	if err != nil {
				// 		fmt.Printf("failed to delete archive %s: %s\n", file, err)
				// 	}
				// }()

				// scp -r
				flags = " -r "
			}
			cfg := &config.ConfigScp{}
			gitar.ReadLastScpConfig(cfg)
			if !last {
				gitar.AskUserInputForScp(cfg)
			}

			//scp -P 2222 go.mod root@192.168.1.100:/tmp/
			//remoteTarget := "/home/" + cfg.User
			remoteTarget := "/tmp/" + filepath.Base(file)
			command := "scp -P " + cfg.Port + " " + flags + file + " " + cfg.User + "@" + cfg.Host + ":" + remoteTarget
			fmt.Println(color.Dim(command))
			if verbose {
				return
			}
			gitar.ExecScp(cfg, file, remoteTarget)
		},
	}
	//scp flags
	scpCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "specify ssh user")
	scpCmd.PersistentFlags().StringVarP(&password, "password", "x", "", "specify ssh user password")
	scpCmd.PersistentFlags().StringVarP(&keyfile, "key", "i", "", "specify ssh private key file")

	// SUBCOMMANDS
	sendCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.Execute()

}
