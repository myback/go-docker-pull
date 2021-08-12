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

func ParseRequestedImage(image string) (*requestedImage, error) {
	pr := &requestedImage{
		tag: defaultTag,
	}

	reqImage := strings.Split(image, "/")
	var ns []string
	switch {
	case len(reqImage) == 1:
		ns = append(ns, officialRepoName)
	case strings.Index(reqImage[0], ".") >= 0, strings.Index(reqImage[0], ":") >= 0:
		pr.registryHost = reqImage[0]
		ns = reqImage[1 : len(reqImage)-1]
	default:
		ns = reqImage[:len(reqImage)-1]
	}

	var imageNameTag []string
	if strings.Index(reqImage[len(reqImage)-1], "@") >= 0 {
		imageNameTag = strings.Split(reqImage[len(reqImage)-1], "@")
	} else {
		imageNameTag = strings.Split(reqImage[len(reqImage)-1], ":")
	}

	ns = append(ns, imageNameTag[0])
	pr.ns = strings.Join(ns, "/")

	switch {
	case len(imageNameTag) == 1:
	case len(imageNameTag) == 2:
		pr.tag = imageNameTag[1]
	default:
		return nil, fmt.Errorf("image format name %s image is invalid", image)
	}

	return pr, nil
}
