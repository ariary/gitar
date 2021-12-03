package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"gitar/pkg/upload"
	"gitar/pkg/utils"
)

// UPLOAD //
//Handler for uploading files
func UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			upload.UploadFile(w, r)
		}
	}
}

//Handler for uploading directory (tar format)
func UploadDirectoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			upload.UntarDirectory(w, r)
		}
	}
}

// ALIAS //
//Handler that output shortcut aimed for the target machines (source it)
func AliasHandler(ip string, port string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//pull
		pullFunc := "pull(){\ncurl -s http://" + ip + ":" + port + "/pull/$1 > $1\n}\n"
		fmt.Fprintf(w, pullFunc)
		//push
		pushFunc := "push(){\ncurl -X POST -F \"file=@$1\" http://" + ip + ":" + port + "/push\n}\n"
		fmt.Fprintf(w, pushFunc)
		//pushr
		pushrFunc := "pushr(){\ntar -cf $1.tar $1 && curl -X POST -F \"file=@$1.tar\" http://" + ip + ":" + port + "/pushr\n}\n"
		fmt.Fprintf(w, pushrFunc)
		//gtree
		gtreeFunc := "gtree(){\ncurl http://" + ip + ":" + port + "/gtree\n}\n"
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

func InitHandlers(directory string, serverIp string, port string) {
	//Upload route
	http.HandleFunc("/push", UploadHandler())

	//Upload directory route
	http.HandleFunc("/pushr", UploadDirectoryHandler())

	//Download route
	http.Handle("/pull/", http.StripPrefix("/pull/", http.FileServer(http.Dir(directory))))

	//Alias endpoint
	http.HandleFunc("/alias", AliasHandler(serverIp, port))

	//Tree endpoint
	http.HandleFunc("/gtree", TreeHandler())
}
