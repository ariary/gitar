package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Check an error and display appropriated message (don't exit or panic)
func Check(e error, msg string) {
	if e != nil {
		if msg != "" {
			fmt.Println(msg)
		}
		fmt.Println(e)
	}
}

// COPY CLIPBOARD // Credits: https://github.com/atotto

//Copy command 'cmd' to clipboard (xclip wrapper)
func Copy(command string) error {
	// Check for xclip installation
	copyCmdArgs := []string{"xclip", "-in", "-selection", "clipboard"}

	_, err := exec.LookPath("xclip")
	Check(err, "xclip was not found")

	// copy
	copyCmd := exec.Command(copyCmdArgs[0], copyCmdArgs[1:]...)
	in, err := copyCmd.StdinPipe()
	Check(err, "Failed redirecting stdin")

	if err := copyCmd.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte(command)); err != nil {
		return err
	}
	if err := in.Close(); err != nil {
		return err
	}
	return copyCmd.Wait()
}

// TAR //

//Untar a "tarball" file  to "target"
func Untar(tarball, target string) error {
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
