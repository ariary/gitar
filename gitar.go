package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func check(e error, msg string) {
	if e != nil {
		if msg != "" {
			fmt.Println(msg)
		}
		fmt.Println(e)
	}
}

//Upload binary file <= 32Mb and return byte content
//Note: upload with curl -X POST -F "file=@[BINARY_FILENAME]" http://[TARGET_IP:PORT]/push
func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	check(err, "Error Retrieving the File")

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check(err, "")
	//write file
	f, err := os.Create(handler.Filename)
	check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check(err, "Error writing to file")
}

//Handler for uploading binary files
func UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			uploadFile(w, r)
		}
	}
}

func main() {
	port := ":" + *flag.String("p", "9237", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	//Upload route
	http.HandleFunc("/push", UploadHandler())

	//Download route
	http.Handle("/pull/", http.StripPrefix("/pull/", http.FileServer(http.Dir(*directory))))

	//Listen
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
