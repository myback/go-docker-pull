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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/docker/image"
	imageV1 "github.com/docker/docker/image/v1"
	"github.com/docker/docker/layer"
	"github.com/docker/docker/pkg/system"
	"github.com/opencontainers/go-digest"
)

const (
	manifestFileName           = "manifest.json"
	legacyLayerFileName        = "layer.tar"
	legacyConfigFileName       = "json"
	legacyVersionFileName      = "VERSION"
	legacyRepositoriesFileName = "repositories"
)

type RegistryClient struct {
	Arch     string
	OS       string
	Login    string
	Password string
	Insecure bool
}

type manifestItem struct {
	Config       string
	RepoTags     []string
	Layers       []string
	Parent       image.ID                                 `json:",omitempty"`
	LayerSources map[layer.DiffID]distribution.Descriptor `json:",omitempty"`
}

func (rc *RegistryClient) Pull(imageReq *requestedImage) error {
	if rc.Insecure {
		imageReq.InsecureRegistry()
	}

	fmt.Printf("%s: Pulling from %s\n", imageReq.tag, imageReq.ns)
	imageReq.insecure = rc.Insecure
	fetcher := &Client{
		Client: &http.Client{},
		Image:  imageReq,
	}
	fetcher.SetCredentials(rc.Login, rc.Password)

	list, err := fetcher.GetManifestList()
	if err != nil {
		if err != ErrEmptyManifestList {
			return err
		}
	}

	imageOS := rc.OS
	imageManifestTag := imageReq.tag
	for _, md := range list.Manifests {
		if md.Platform.Architecture == rc.Arch && md.Platform.OS == rc.OS {
			imageOS = md.Platform.OS
			imageManifestTag = md.Digest.String()
			break
		}
	}

	imageManifest, contentDigest, err := fetcher.GetManifest(imageManifestTag)
	if err != nil {
		return err
	}

	imageManifestFilename := imageManifest.Config.Digest.Hex() + ".json"

	tmpDir, err := imageReq.TempDirCreate()
	if err != nil {
		return err
	}

	resp, err := fetcher.GetBlob(imageManifest.Config.Digest, "", 0)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	imageRepoBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	imageManifestFilepath := filepath.Join(tmpDir, imageManifestFilename)
	f, err := os.Create(imageManifestFilepath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(imageRepoBytes)

	imageConfig := image.Image{}
	if err := json.Unmarshal(imageRepoBytes, &imageConfig); err != nil {
		return err
	}

	imageRepo := imageReq.ns
	if strings.HasPrefix(imageReq.ns, officialRepoName+"/") && imageReq.registryHost == "" {
		imageRepo = strings.Replace(imageReq.ns, officialRepoName+"/", "", 1)
	}

	if err := system.Chtimes(imageManifestFilepath, imageConfig.Created, imageConfig.Created); err != nil {
		return err
	}

	var manifest []manifestItem
	newImageManifest := manifestItem{
		Config:   imageManifestFilename,
		RepoTags: []string{imageRepo + ":" + imageReq.tag},
	}

	var parentId digest.Digest
	for i, diffId := range imageConfig.RootFS.DiffIDs {
		v1Img := image.V1Image{
			Created: time.Unix(0, 0).UTC(),
		}

		if i == len(imageConfig.RootFS.DiffIDs)-1 {
			v1Img = imageConfig.V1Image
		}
		rootFS := *imageConfig.RootFS
		rootFS.DiffIDs = rootFS.DiffIDs[:i+1]
		v1ID, err := imageV1.CreateID(v1Img, rootFS.ChainID(), parentId)
		if err != nil {
			return err
		}

		if parentId != "" {
			v1Img.Parent = parentId.Hex()
		}
		parentId = v1ID
		v1Img.ID = v1ID.Hex()
		v1Img.OS = imageOS
		newImageManifest.Layers = append(newImageManifest.Layers, filepath.Join(v1Img.ID, legacyLayerFileName))

		if err := fetcher.GetLayer(tmpDir, digest.Digest(diffId), imageManifest.Layers[i], v1Img, imageConfig.Created.UTC()); err != nil {
			return err
		}
	}

	manifest = append(manifest, newImageManifest)

	if err := SaveToJson(filepath.Join(tmpDir, manifestFileName), manifest); err != nil {
		return err
	}

	if err := SaveToJson(filepath.Join(tmpDir, legacyRepositoriesFileName), map[string]map[string]string{
		imageRepo: {imageReq.tag: parentId.Hex()},
	}); err != nil {
		return err
	}

	fmt.Println("Digest:", contentDigest)

	return chtimes(tmpDir, []string{manifestFileName, legacyRepositoriesFileName}, time.Unix(0, 0))
}
