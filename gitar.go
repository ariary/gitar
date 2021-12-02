package main

import (
	"archive/tar"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func check(e error, msg string) {
	if e != nil {
		if msg != "" {
			fmt.Println(msg)
		}
		fmt.Println(e)
	}
}

// UPLOAD //

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

	//write file
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check(err, "")

	f, err := os.Create(handler.Filename)
	check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check(err, "Error writing to file")
}

//Handler for uploading files
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

// UPLOAD DIRECTORY//

//untar a "tarball" file  to "target"
func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

//Untar directory from http request (dl it, untar it, remove it)
func untarDirectory(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	check(err, "Error Retrieving the File")

	defer file.Close()

	filename := handler.Filename[:strings.LastIndex(handler.Filename, ".")] //handler.Filename - .tar
	fmt.Printf("Uploaded Directory: %+v\n", filename)

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check(err, "")
	//write file
	f, err := os.Create(handler.Filename)
	check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check(err, "Error writing to file")
	untar(handler.Filename, filename)
	check(os.Remove(handler.Filename), "Error while remove directory tar")
}

//Handler for uploading directory (tar format)
func UploadDirectoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			untarDirectory(w, r)
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
		check(err, "Error while executing tree command")
		fmt.Fprintf(w, string(out))
	}
}

func main() {
	serverIp := flag.String("e", "127.0.0.1", "server external reachable ip")
	port := flag.String("p", "9237", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	//Upload route
	http.HandleFunc("/push", UploadHandler())

	//Upload directory route
	http.HandleFunc("/pushr", UploadDirectoryHandler())

	//Download route
	http.Handle("/pull/", http.StripPrefix("/pull/", http.FileServer(http.Dir(*directory))))

	//Alias endpoint
	http.HandleFunc("/alias", AliasHandler(*serverIp, *port))

	//Tree endpoint
	http.HandleFunc("/gtree", TreeHandler())

	//Set up messages
	fmt.Println("On remote:")
	fmt.Println("curl -s http://" + *serverIp + ":" + *port + "/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias")

	//Listen
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
