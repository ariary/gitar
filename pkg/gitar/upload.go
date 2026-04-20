package gitar

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/gitar/pkg/utils"
	"github.com/ariary/go-utils/pkg/check"
	"github.com/ariary/go-utils/pkg/color"
)

// UPLOAD //

//Upload binary file <= 32Mb and return byte content
//Note: upload with curl -X POST -F "file=@[BINARY_FILENAME]" http://[TARGET_IP:PORT]/push
//Note: also handle multipart form for fun
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

	upFilename := filepath.Join(upDir, filepath.Base(handler.Filename))
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

	baseName := filepath.Base(handler.Filename)
	filename := baseName[:strings.LastIndex(baseName, ".")] // strip .tar
	fmt.Printf("Upload Directory: %+v\n", color.Bold(filename))
	filename = filepath.Join(upDir, filename)

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	check.Check(err, "")
	//write file
	upFilename := filepath.Join(upDir, baseName)
	f, err := os.Create(upFilename)
	check.Check(err, "Error creating file")

	defer f.Close()

	_, err = f.Write(buf.Bytes())
	check.Check(err, "Error writing to file")
	utils.Untar(upFilename, filename)
	check.Check(os.Remove(upFilename), "Error while remove directory tar")
}

// UploadZipHandler wraps UploadZipDirectory as an http.HandlerFunc
func UploadZipHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		remote := "(" + r.RemoteAddr + ")"
		fmt.Print(color.Teal(remote), " ")
		switch r.Method {
		case "GET":
			fmt.Println("Get request")
			http.Error(w, "GET Bad request - Only POST accepted!", 400)
		case "POST":
			UploadZipDirectory(cfg.UploadDir, w, r)
		}
	}
}

// UploadZipDirectory receives a .zip archive and extracts it into upDir.
// Used by the Windows PowerShell pushr alias.
func UploadZipDirectory(upDir string, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	check.Check(err, "Error Retrieving the File")
	defer file.Close()

	dirName := strings.TrimSuffix(filepath.Base(handler.Filename), ".zip")
	fmt.Printf("Upload Directory (zip): %+v\n", color.Bold(dirName))

	// Write zip to a temp file so zip.OpenReader can seek it
	tmpFile, err := os.CreateTemp("", "gitar-*.zip")
	check.Check(err, "Error creating temp file for zip")
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	_, err = io.Copy(tmpFile, file)
	check.Check(err, "Error writing zip to temp file")
	tmpFile.Close()

	destDir := filepath.Join(upDir, dirName)
	err = unzipDir(tmpPath, destDir)
	check.Check(err, "Error extracting zip archive")
}

// unzipDir extracts src zip archive into dest directory.
// Returns an error if any entry would escape dest (path traversal guard).
func unzipDir(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	dest = filepath.Clean(dest)
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, dest+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry %q escapes destination directory", f.Name)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
