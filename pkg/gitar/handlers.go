package gitar

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ariary/gitar/pkg/config"

	"github.com/ariary/go-utils/pkg/check"
	"github.com/ariary/go-utils/pkg/color"
)

// UPLOAD //
//Handler for uploading files
func UploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		remote := "(" + r.RemoteAddr + ")"
		fmt.Print(color.Teal(remote), " ")
		switch r.Method {
		case "GET":
			fmt.Println("Get request")
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			UploadFile(cfg.UploadDir, w, r)
		}
	}
}

//Handler for uploading directory (tar format)
func UploadDirectoryHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		remote := "(" + r.RemoteAddr + ")"
		fmt.Print(color.Teal(remote), " ")
		switch r.Method {
		case "GET":
			fmt.Println("Get request")
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			UntarDirectory(cfg.UploadDir, w, r)
		}
	}
}

//Handler for uploading directory (tar format)
func DownloadHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remote := "(" + r.RemoteAddr + ")"
		file := strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
		fmt.Println(color.Green(remote), "Download file:", color.Bold(file))
		h.ServeHTTP(w, r)
	})
}

//Handler for bidirectional exchange (target has previously set up a webhook to continuously pull this repo. Once downloaded, file is deleted)
func bidirectionalHandler(h http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remote := "(" + r.RemoteAddr + ")"
		h.ServeHTTP(w, r)
		file := strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
		if file != "" {
			fmt.Println(color.Green(remote), "Download file", color.Italic("(push to remote)"), ":", color.Bold(file))
			file = cfg.BidirectionalDir + "/" + file
			err := exec.Command("rm", file).Run()
			check.Check(err, "Error while removing "+file)
		}
	})
}

// ALIAS //
//Handler that output shortcut aimed for the target machines (source it). It is for linux machines
func AliasHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		url := cfg.Url

		//pull
		pullFunc := "pull(){\nFILE=$(echo $1| rev | cut -d\"/\" -f 1 | rev)\ncurl -s " + url + "/pull/$1 > $FILE\n}\n"
		fmt.Fprintf(w, pullFunc)

		//pullr
		statusFunc := "status(){\ncurl -s -o /dev/null -w \"%s{http_code}\" " + url + "/pull/$1\n}\n"
		fmt.Fprintf(w, statusFunc, "%")

		getAllFillesFunc := "getFiles(){\ncurl -L -s " + url + "/pull/$1 | grep \"<a\" | cut -d \"\\\"\" -f 2\n}\n"
		fmt.Fprintf(w, getAllFillesFunc)

		isDirFunc := "isDir(){\n[[ \"$1\" == */ ]]\n}\n"
		fmt.Fprintf(w, isDirFunc)

		pullrFunc := `pullr(){
			STATUS=$(status $1)
			if [ $STATUS -eq 301  ]
				mkdir -p $1
			then
				FILES=$(getFiles "$1")
			fi
			#fix zsh bug
			local IFS=$'\n'
			if [ $ZSH_VERSION ]; then
			  setopt sh_word_split
			fi
			
			for value in $FILES
			do
				if isDir $value
				then
					value=${value::-1}
				fi
				file="$1/$value"
				STATUS=$(status $file)
				if [ $STATUS -eq 301  ]
				then
					# echo "$file"
					pullr $file
				else
					# echo "$file"
					pull $file
					mv $value $file
				fi
			done
			}
			`
		//pullrFunc := "pullr(){\nSTATUS=$(status $1)\nif [ $STATUS -eq 301  ]\nmkdir $1\nthen\nFILES=$(getFiles \"$1\")\nfor value in $FILES\ndo\nif isDir $value\nthen\nvalue=${value::-1}\nfi\nfile=\"$1/$value\"\nSTATUS=$(status $file)\nif [ $STATUS -eq 301  ]\nthen\npullr $file\nelse\npull $file\nfi\ndone\nfi\n}\n"
		fmt.Fprintf(w, pullrFunc)

		//push
		pushFunc := "push(){\ncurl -X POST -F \"file=@$1\" " + url + "/push\n}\n"
		fmt.Fprintf(w, pushFunc)

		//pushr
		pushrFunc := "pushr(){\ntar -cf $1.tar $1 && curl -X POST -F \"file=@$1.tar\" " + url + "/pushr && rm $1.tar\n}\n"
		fmt.Fprintf(w, pushrFunc)

		//receive
		if cfg.BidirectionalDir != "" {
			receiveFunc := `
			receive(){
				while true
				do
					FILES=$(curl -s ` + url + `/bidirectional -L | grep "<a" | cut -d "\"" -f 2)
					#fix zsh bug
					local IFS=$'\n'
					if [ $ZSH_VERSION ]; then
					setopt sh_word_split
					fi
					
					for value in $FILES
					do
						curl -s ` + url + `/bidirectional/$value > $value
					done
					sleep 5
				done
			}
			(&>/dev/null receive &)
			`
			fmt.Fprintf(w, receiveFunc)
		}

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
			//withdraw dir prefix to be consistent with handler endpoint
			if dir != "." {
				path = strings.Replace(path, dir+"/", "", 1)
			}

			if !info.IsDir() {
				files = files + " " + path
			} else {
				directories = directories + " " + path
			}
			return nil
		})
	check.Check(err, "Failed retrieving files for completion")

	//create completion lines

	// pushC := "complete -A push"
	// pushrC := "complete -A pushr"
	pullC := "complete -W \"" + files + "\" pull"
	pullrC := "complete -W \"" + directories + "\" pullr"
	completionLines := []string{pullC, pullrC}

	for i := 0; i < len(completionLines); i++ {
		completions += completionLines[i] + "\n"
	}

	return completions
}

//Handler that output shortcut aimed for the target machines (source it). It is for windows machine with powershell
// An alternative to Invoke-WebRequest can be Invoke-RestMethod
func AliasWindowsPS(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		url := cfg.Url

		//pull
		pullFunc := "function pull([string]$file){\nInvoke-WebRequest " + url + "/pull/$file -OutFile $file\n}\n"
		fmt.Fprintf(w, pullFunc)

		//push
		pushFunc := "function push([string]$file){\n$Uri = '" + url + "/push'\n$Form = @{file = Get-Item -Path $file}\n$Result = Invoke-WebRequest -Uri $Uri -Method Post -Form $Form\n}\n"
		fmt.Fprintf(w, pushFunc)

		//gtree
		gtreeFunc := "function gtree(){\n(Invoke-WebRequest -Uri " + url + "/gtree).Content\n}\n"
		fmt.Fprintf(w, gtreeFunc)

		//pushr
		//pullr
		//Completion

	}
}

//Handler that output shortcut aimed for the target machines (source it). It is for windows machine in cmd.exe
func AliasWindowsCmdHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// echoFunc := "function echy([string]$one) {echo $one}"
		// fmt.Fprintf(w, echoFunc)

		// url := cfg.Url

		// //pull
		// pullFunc := "function pull([string]$file){\n(curl " + url + "/pull/$file).Content > $file\n}\n"
		// fmt.Fprintf(w, pullFunc)
	}
}

// TREE //
//Handler that print the tree of the file server
func TreeHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := exec.Command("tree", cfg.DownloadDir).Output()
		check.Check(err, "Error while executing tree command")
		fmt.Fprintf(w, string(out))
	}
}

// INIT //

func InitHandlers(cfg *config.Config) {
	//Upload route
	http.HandleFunc("/"+cfg.Secret+"/push", UploadHandler(cfg))

	//Upload directory route
	http.HandleFunc("/"+cfg.Secret+"/pushr", UploadDirectoryHandler(cfg))

	//Download route
	//http.Handle("/pull/", http.StripPrefix("/pull/", http.FileServer(http.Dir(cfg.DownloadDir))))
	http.Handle("/"+cfg.Secret+"/pull/", DownloadHandler(http.StripPrefix("/"+cfg.Secret+"/pull/", http.FileServer(http.Dir(cfg.DownloadDir)))))

	//Alias endpoint
	http.HandleFunc("/"+cfg.Secret+"/alias", AliasHandler(cfg))
	http.HandleFunc("/"+cfg.Secret+"/aliaswinps", AliasWindowsPS(cfg))
	http.HandleFunc("/"+cfg.Secret+"/aliaswincmd", AliasWindowsCmdHandler(cfg))

	//"Bidirectional" endpoint
	http.Handle("/"+cfg.Secret+"/bidirectional/", bidirectionalHandler(http.StripPrefix("/"+cfg.Secret+"/bidirectional/", http.FileServer(http.Dir(cfg.BidirectionalDir))), cfg))

	//Tree endpoint
	http.HandleFunc("/"+cfg.Secret+"/gtree", TreeHandler(cfg))
}
