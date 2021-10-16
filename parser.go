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

package dockerPull

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/registry"
)

const (
	officialRepoName = "library"
	defaultTag       = "latest"
)

type requestedImage struct {
	insecure     bool
	registryHost string
	ns           string
	tag          string
	tempDir      string
}

func (ri *requestedImage) Url(paths ...string) string {
	u := registry.DefaultV2Registry

	if ri.insecure {
		u.Scheme = "http"
	}
	if ri.registryHost != "" {
		u.Host = ri.registryHost
	}
	u.Path = path.Join("v2", ri.ns)
	u.Path = path.Join(append([]string{u.Path}, paths...)...)

	return u.String()
}

func (ri *requestedImage) ManifestUrl(tag string) string {
	return ri.Url("manifests", tag)
}

func (ri *requestedImage) BlobUrl(tag string) string {
	return ri.Url("blobs", tag)
}

func (ri *requestedImage) Tag() string {
	return ri.tag
}

func (ri *requestedImage) InsecureRegistry() {
	ri.insecure = false
}

func (ri *requestedImage) OutputImageName() string {
	return fmt.Sprintf("%s_%s.tar", strings.ReplaceAll(ri.ns, "/", "_"),
		strings.ReplaceAll(ri.tag, "-", "_"))
}

func (ri *requestedImage) TempDir() string {
	return ri.tempDir
}

func (ri *requestedImage) TempDirCreate() (string, error) {
	ri.tempDir = fmt.Sprintf("%s_%s.tmp", strings.ReplaceAll(ri.ns, "/", "_"),
		strings.ReplaceAll(ri.tag, "-", "_"))

	return ri.tempDir, os.MkdirAll(ri.tempDir, os.ModePerm)
}

func ParseRequestedImage(s string) *requestedImage {
	ri := &requestedImage{
		tag: defaultTag,
	}

	idx := strings.IndexByte(s, '/')
	if idx > -1 && strings.IndexAny(s[:idx], ".:") > -1 {
		ri.registryHost = s[:idx]
		s = s[idx+1:]
	}

	idx = strings.IndexAny(s, "@:")
	if idx > -1 {
		ri.tag = s[idx+1:]
		s = s[:idx]
	}

	idx = strings.IndexByte(s, '/')
	if idx == -1 && ri.registryHost == "" {
		s = officialRepoName + "/" + s
	}

	ri.ns = s

	return ri
}
