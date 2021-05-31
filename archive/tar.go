/*
Copyright © 2021 myback.space <git@myback.space>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	modeISDIR  = 040000  // Directory
	modeISFIFO = 010000  // FIFO
	modeISREG  = 0100000 // Regular file
	modeISLNK  = 0120000 // Symbolic link
	modeISBLK  = 060000  // Block special file
	modeISCHR  = 020000  // Character special file
	modeISSOCK = 0140000 // Socket
)

func chmodTarEntry(perm os.FileMode) os.FileMode {
	return perm // noop for unix as golang APIs provide perm bits correctly
}

func fillGo18FileTypeBits(mode int64, fi os.FileInfo) int64 {
	fm := fi.Mode()
	switch {
	case fm.IsRegular():
		mode |= modeISREG
	case fi.IsDir():
		mode |= modeISDIR
	case fm&os.ModeSymlink != 0:
		mode |= modeISLNK
	case fm&os.ModeDevice != 0:
		if fm&os.ModeCharDevice != 0 {
			mode |= modeISCHR
		} else {
			mode |= modeISBLK
		}
	case fm&os.ModeNamedPipe != 0:
		mode |= modeISFIFO
	case fm&os.ModeSocket != 0:
		mode |= modeISSOCK
	}
	return mode
}

func Tar(srcPath, dst string) error {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return TarStream(f, srcPath)
}

func TarStream(dst io.Writer, srcPath string) error {
	tarWriter := tar.NewWriter(dst)
	defer tarWriter.Close()

	return filepath.Walk(srcPath, func(filePath string, fi os.FileInfo, err error) error {
		headers, err := tar.FileInfoHeader(fi, filePath)
		if err != nil {
			return err
		}
		relFilePath, err := filepath.Rel(srcPath, filePath)
		if err != nil || relFilePath == "." && fi.IsDir() {
			// Error getting relative path OR we are looking
			// at the source directory path. Skip in both situations.
			return nil
		}

		relFilePath = filepath.ToSlash(relFilePath)

		if fi.IsDir() {
			relFilePath += "/"
		}

		headers.ModTime = headers.ModTime.Truncate(time.Second)
		headers.Uname = ""
		headers.Gname = ""
		headers.Uid = 0
		headers.Gid = 0
		headers.Name = relFilePath
		headers.Mode = fillGo18FileTypeBits(int64(chmodTarEntry(os.FileMode(headers.Mode))), fi)

		if err := tarWriter.WriteHeader(headers); err != nil {
			return fmt.Errorf("write headers failed: %s", err)
		}

		if !fi.IsDir() {
			data, err := os.Open(filePath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				data.Close()
				return err
			}
			data.Close()
		}
		return nil
	})
}

func Untar(dst string, src io.Reader) error {
	tarReader := tar.NewReader(src)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(dst, header.Name), os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := untarCreateFile(filepath.Join(dst, header.Name), tarReader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("extract tar: uknown type: %s in %s", string(header.Typeflag), header.Name)
		}
	}
}

func untarCreateFile(path string, reader *tar.Reader) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)

	return err
}
