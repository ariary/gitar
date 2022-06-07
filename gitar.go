package main

import (
	"fmt"
	"os"

	"github.com/ariary/gitar/pkg/gitar"
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

	//CMD OUTGOING/LIGHT
	var cmdSend = &cobra.Command{
		Use:   "send",
		Short: "directly send a file to a target using different options",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("file:", args[0])
		},
	}

	// FLAGS
	//token,pageid,p,external,shell,delay,server,config-from-page
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

	rootCmd.AddCommand(cmdSend)
	rootCmd.Execute()

}
