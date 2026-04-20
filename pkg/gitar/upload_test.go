package gitar

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadZipDirectory(t *testing.T) {
	// Build a zip with two files: file1.txt and sub/file2.txt
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f1, _ := zw.Create("file1.txt")
	f1.Write([]byte("hello"))
	f2, _ := zw.Create("sub/file2.txt")
	f2.Write([]byte("world"))
	zw.Close()

	// Create multipart POST body
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "testdir.zip")
	fw.Write(zipBuf.Bytes())
	mw.Close()

	upDir := t.TempDir() + "/"
	req := httptest.NewRequest("POST", "/pushrzip", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	UploadZipDirectory(upDir, w, req)

	if _, err := os.Stat(filepath.Join(upDir, "testdir", "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt was not extracted into testdir/")
	}
	if _, err := os.Stat(filepath.Join(upDir, "testdir", "sub", "file2.txt")); os.IsNotExist(err) {
		t.Error("sub/file2.txt was not extracted into testdir/sub/")
	}
}

func TestUnzipDirPathTraversal(t *testing.T) {
	// Build a zip with a path-traversal entry
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("../../evil.txt")
	f.Write([]byte("bad"))
	zw.Close()

	tmpZip, _ := os.CreateTemp("", "*.zip")
	tmpZip.Write(zipBuf.Bytes())
	tmpZip.Close()
	defer os.Remove(tmpZip.Name())

	dest := t.TempDir()
	err := unzipDir(tmpZip.Name(), dest)
	if err == nil {
		t.Error("expected error for path-traversal zip entry, got nil")
	}
}
