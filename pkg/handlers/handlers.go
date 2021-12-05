package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"gitar/pkg/config"
	"gitar/pkg/upload"
	"gitar/pkg/utils"
)

// UPLOAD //
//Handler for uploading files
func UploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			upload.UploadFile(cfg.UploadDir, w, r)
		}
	}
}

//Handler for uploading directory (tar format)
func UploadDirectoryHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			upload.UntarDirectory(cfg.UploadDir, w, r)
		}
	}
}

// ALIAS //
//Handler that output shortcut aimed for the target machines (source it)
func AliasHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := cfg.ServerIP
		port := cfg.Port
		var protocol string
		if cfg.Tls {
			protocol = "-k https://"
		} else {
			protocol = "http://"
		}
		url := protocol + ip + ":" + port

		//pull
		pullFunc := "pull(){\ncurl -s " + url + "/pull/$1 > $1\n}\n"
		fmt.Fprintf(w, pullFunc)
		//push
		pushFunc := "push(){\ncurl -X POST -F \"file=@$1\" " + url + "/push\n}\n"
		fmt.Fprintf(w, pushFunc)
		//pushr
		pushrFunc := "pushr(){\ntar -cf $1.tar $1 && curl -X POST -F \"file=@$1.tar\" " + url + "/pushr && rm $1.tar\n}\n"
		fmt.Fprintf(w, pushrFunc)
		//gtree
		gtreeFunc := "gtree(){\ncurl " + url + "/gtree\n}\n"
		fmt.Fprintf(w, gtreeFunc)
	}
}

// TREE //
//Handler that print the tree of the file server
func TreeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := exec.Command("tree").Output()
		utils.Check(err, "Error while executing tree command")
		fmt.Fprintf(w, string(out))
	}
}

// INIT //

func InitHandlers(cfg *config.Config) {
	//Upload route
	http.HandleFunc("/push", UploadHandler(cfg))

	//Upload directory route
	http.HandleFunc("/pushr", UploadDirectoryHandler(cfg))

	//Download route
	http.Handle("/pull/", http.StripPrefix("/pull/", http.FileServer(http.Dir(cfg.DownloadDir))))

	//Alias endpoint
	http.HandleFunc("/alias", AliasHandler(cfg))

	//Tree endpoint
	http.HandleFunc("/gtree", TreeHandler())
}
