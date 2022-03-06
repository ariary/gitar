package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ariary/gitar/pkg/utils"
	"github.com/ariary/go-utils/pkg/check"
	"github.com/ariary/go-utils/pkg/color"
)

// UPLOAD //

//Upload binary file <= 32Mb and return byte content
//Note: upload with curl -X POST -F "file=@[BINARY_FILENAME]" http://[TARGET_IP:PORT]/push
func UploadFile(upDir string, w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	check.Check(err, "Error Retrieving the File")

	defer file.Close()
	fmt.Printf("Upload File: %+v\n", color.Bold(handler.Filename))

	//write file
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check.Check(err, "")

	upFilename := upDir + handler.Filename
	f, err := os.Create(upFilename)
	check.Check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check.Check(err, "Error writing to file")
}

//Untar directory from http request (dl it, untar it, remove it)
func UntarDirectory(upDir string, w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(32 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	check.Check(err, "Error Retrieving the File")

	defer file.Close()

	filename := handler.Filename[:strings.LastIndex(handler.Filename, ".")] //handler.Filename - .tar
	fmt.Printf("Upload Directory: %+v\n", color.Bold(filename))
	filename = upDir + filename

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check.Check(err, "")
	//write file
	upFilename := upDir + handler.Filename
	f, err := os.Create(upFilename)
	check.Check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check.Check(err, "Error writing to file")
	utils.Untar(upFilename, filename)
	check.Check(os.Remove(upFilename), "Error while remove directory tar")
}
