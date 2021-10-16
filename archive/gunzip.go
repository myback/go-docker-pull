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
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
)

type Gzip struct {
	dst, src string
}

func (a *Gzip) GunZip(multiWriter ...io.Writer) (int64, error) {
	file, err := os.Open(a.src)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, err
	}
	defer gzipReader.Close()

	unArchFile, err := os.Create(a.dst)
	if err != nil {
		return 0, err
	}
	defer unArchFile.Close()

	return io.Copy(io.MultiWriter(append([]io.Writer{unArchFile}, multiWriter...)...), gzipReader)
}

func (a *Gzip) GetUnarchSize() (uint32, error) {
	file, err := os.Open(a.src)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	leSize := make([]byte, 4)
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	if _, err := file.ReadAt(leSize, stat.Size()-4); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(leSize), nil
}

func NewGzip(dst, src string) *Gzip {
	return &Gzip{
		dst: dst,
		src: src,
	}
}
