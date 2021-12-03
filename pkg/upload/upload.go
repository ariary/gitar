package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gitar/pkg/utils"
)

// UPLOAD //

//Upload binary file <= 32Mb and return byte content
//Note: upload with curl -X POST -F "file=@[BINARY_FILENAME]" http://[TARGET_IP:PORT]/push
func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	utils.Check(err, "Error Retrieving the File")

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)

	//write file
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	utils.Check(err, "")

	f, err := os.Create(handler.Filename)
	utils.Check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	utils.Check(err, "Error writing to file")
}

//Untar directory from http request (dl it, untar it, remove it)
func UntarDirectory(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	utils.Check(err, "Error Retrieving the File")

	defer file.Close()

	filename := handler.Filename[:strings.LastIndex(handler.Filename, ".")] //handler.Filename - .tar
	fmt.Printf("Uploaded Directory: %+v\n", filename)

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	utils.Check(err, "")
	//write file
	f, err := os.Create(handler.Filename)
	utils.Check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	utils.Check(err, "Error writing to file")
	utils.Untar(handler.Filename, filename)
	utils.Check(os.Remove(handler.Filename), "Error while remove directory tar")
}