package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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
		addr := cfg.ServerIP
		port := cfg.Port
		var protocol string
		if cfg.Tls {
			protocol = "-k https://"
		} else {
			protocol = "http://"
		}
		url := protocol + addr + ":" + port
		if cfg.AliasUrl != "" {
			url = cfg.AliasUrl
		}

		//pull
		pullFunc := "pull(){\ncurl -s " + url + "/pull/$1 > $1\n}\n"
		fmt.Fprintf(w, pullFunc)

		//pullr
		statusFunc := "status(){\ncurl -s -o /dev/null -w \"%s{http_code}\" " + url + "/pull/$1\n}\n"
		fmt.Fprintf(w, statusFunc, "%")

		getAllFillesFunc := "getFiles(){\ncurl -L -s http://127.0.0.1:9237/pull/$1 | grep \"<a\" | cut -d \"\\\"\" -f 2\n}\n"
		fmt.Fprintf(w, getAllFillesFunc)

		isDirFunc := "isDir(){\n[[ \"$1\" == */ ]]\n}\n"
		fmt.Fprintf(w, isDirFunc)

		pullrFunc := "pullr(){\nSTATUS=$(status $1)\nif [ $STATUS -eq 301  ]\nmkdir $1\nthen\nFILES=$(getFiles \"$1\")\nfor value in $FILES\ndo\nif isDir $value\nthen\nvalue=${value::-1}\nfi\nfile=\"$1/$value\"\nSTATUS=$(status $file)\nif [ $STATUS -eq 301  ]\nthen\npullr $file\nelse\npull $file\nfi\ndone\nfi\n}\n"
		fmt.Fprintf(w, pullrFunc)

		//push
		pushFunc := "push(){\ncurl -X POST -F \"file=@$1\" " + url + "/push\n}\n"
		fmt.Fprintf(w, pushFunc)

		//pushr
		pushrFunc := "pushr(){\ntar -cf $1.tar $1 && curl -X POST -F \"file=@$1.tar\" " + url + "/pushr && rm $1.tar\n}\n"
		fmt.Fprintf(w, pushrFunc)

		//gtree
		gtreeFunc := "gtree(){\ncurl " + url + "/gtree\n}\n"
		fmt.Fprintf(w, gtreeFunc)

		//Completion
		if cfg.Completion {
			fmt.Fprintf(w, getCompletion(cfg.DownloadDir))
		}
	}
}

//Return the completion command to source
func getCompletion(dir string) (completions string) {
	//retrieve all file & directory of dir
	var files string
	var directories string
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = files + " " + path
			} else {
				directories = directories + " " + path
			}
			return nil
		})
	utils.Check(err, "Failed retrieving files for completion")

	//create completion lines

	// pushC := "complete -A push"
	// pushrC := "complete -A pushr"
	pullC := "complete -W \"" + files + "\" pull"
	pullrC := "complete -W \"" + directories + "\" pullr"
	completionLines := []string{pullC, pullrC}

	for i := 0; i < len(completionLines); i++ {
		fmt.Println(completionLines[i])
		completions += completionLines[i] + "\n"
	}

	return completions
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
