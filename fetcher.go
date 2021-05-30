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
	"github.com/docker/distribution"
	"github.com/docker/docker/image"
	"github.com/myback/go-docker-pull/archive"
	"github.com/myback/go-docker-pull/progressbar"
	"github.com/opencontainers/go-digest"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
)

var (
	ErrImageNotFound     = fmt.Errorf("pull access denied, repository does not exist or may require login and password")
	ErrEmptyManifestList = fmt.Errorf("empty manifest list")
)

type jwtToken struct {
	Token     string    `json:"token"`
	ExpiresIn int       `json:"expires_in"`
	IssuedAt  time.Time `json:"issued_at"`
}

type Client struct {
	*http.Client
	Image               *requestedImage
	token               *jwtToken
	login, password, UA string
}

func (c *Client) SetCredentials(login, password string) {
	c.login = login
	c.password = password
	c.token = nil
}

func (c *Client) NewGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.UA == "" {
		c.UA = "docker-pull"
	}

	req.Header.Set("User-Agent", c.UA)
	return req, nil
}

func (c *Client) getToken(url string) error {
	req, err := c.NewGetRequest(url)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	realm := WWWAuthenticateParse(resp.Header.Get("www-authenticate"))
	u, err := realm.Url()
	if err != nil {
		return err
	}

	reqToken, err := c.NewGetRequest(u)
	if err != nil {
		return err
	}

	if c.login != "" {
		reqToken.SetBasicAuth(c.login, c.password)
	}

	respToken, err := c.Do(reqToken)
	if err != nil {
		return err
	}
	defer respToken.Body.Close()

	if c.token == nil {
		c.token = &jwtToken{}
	}

	if err := json.NewDecoder(respToken.Body).Decode(c.token); err != nil {
		return err
	}

	return nil
}

func (c *Client) get(url string, headers http.Header) (*http.Response, error) {
	// || time.Now().After(c.token.IssuedAt.Add(time.Duration(c.token.ExpiresIn-1)*time.Second))
	if c.token == nil {
		if err := c.getToken(url); err != nil {
			return nil, err
		}
	}

	req, err := c.NewGetRequest(url)
	if err != nil {
		return nil, err
	}

	req.Header = headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Token))

	resp, err := c.Do(req)
	if err == nil {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if err := c.getToken(url); err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Token))
			resp, err = c.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
				return nil, ErrImageNotFound
			}
		}
	}

	return resp, err
}

func (c *Client) GetManifestList() (*manifestlist.ManifestList, error) {
	list := &manifestlist.ManifestList{}

	hdr := http.Header{}
	hdr.Set("Accept", manifestlist.MediaTypeManifestList)
	resp, err := c.get(c.Image.ManifestUrl(c.Image.tag), hdr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(list); err != nil {
		return nil, err
	}

	if list.SchemaVersion != 2 || len(list.Manifests) == 0 {
		return list, ErrEmptyManifestList
	}

	return list, nil
}

func (c *Client) GetManifest(tag string) (*schema2.Manifest, string, error) {
	manifest := &schema2.Manifest{}

	hdr := http.Header{}
	hdr.Set("Accept", schema2.MediaTypeManifest)
	resp, err := c.get(c.Image.ManifestUrl(tag), hdr)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	contentDigest := resp.Header.Get("Docker-Content-Digest")

	if err := json.NewDecoder(resp.Body).Decode(manifest); err != nil {
		return nil, "", err
	}

	return manifest, contentDigest, nil
}

func (c *Client) GetBlob(tag digest.Digest, mediaTypeLayer string, resume int64) (*http.Response, error) {
	hdr := http.Header{}
	if mediaTypeLayer != "" {
		hdr.Set("Accept", mediaTypeLayer)
	}

	if resume > 0 {
		hdr.Set("Range", fmt.Sprintf("bytes=%d-", resume))
	}

	resp, err := c.get(c.Image.BlobUrl(tag.String()), hdr)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code [%d]: error: \"%s\"", resp.StatusCode, b)
	}

	return resp, err
}

func (c *Client) GetLayer(dir string, diffId digest.Digest, layerDesc distribution.Descriptor, legacyImg image.V1Image, created time.Time) error {
	legacyFilesList := []string{"", legacyVersionFileName, legacyConfigFileName, legacyLayerFileName}
	outDir := filepath.Join(dir, legacyImg.ID)

	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(outDir, legacyVersionFileName), []byte("1.0"), 0644); err != nil {
		return err
	}

	legacyJson, err := json.Marshal(legacyImg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(outDir, legacyConfigFileName), legacyJson, 0644); err != nil {
		return err
	}

	layerFilePath := filepath.Join(outDir, legacyLayerFileName)
	tmpLayer := layerFilePath + ".gz"
	shortLayerTag := layerDesc.Digest.Hex()[:12]

	bar := progressbar.NewProgressBar(50)
	defer bar.Close()

	if _, err := os.Stat(layerFilePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		ok, err := FileHashEqual(layerFilePath, diffId.Hex())
		if err != nil {
			return err
		}

		if ok {
			bar.SetDescription(fmt.Sprintf("%s: %s ", shortLayerTag, "Pull complete"))
			bar.Flush()

			return chtimes(outDir, legacyFilesList, created)
		}
	}

	var resume int64
	if fInfo, err := os.Stat(tmpLayer); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		ok, err := FileHashEqual(tmpLayer, layerDesc.Digest.Hex())
		if err != nil {
			return err
		}

		if ok {
			return gunzipLayer(layerFilePath, tmpLayer, shortLayerTag, bar)
		}

		resume = fInfo.Size()
	}

	resp, err := c.GetBlob(layerDesc.Digest, layerDesc.MediaType, resume)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outputFile, err := os.OpenFile(tmpLayer, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	bar.ContentLength(resp.ContentLength)
	bar.SetDescription(fmt.Sprintf("%s: %s ", shortLayerTag, "Downloading"))

	buff := make([]byte, 131072)
	if _, err := io.CopyBuffer(io.MultiWriter(outputFile, bar), resp.Body, buff); err != nil {
		return err
	}

	if err := gunzipLayer(layerFilePath, tmpLayer, shortLayerTag, bar); err != nil {
		return err
	}

	bar.SetDescription(fmt.Sprintf("%s: %s ", shortLayerTag, "Pull complete"))
	bar.Flush()

	return chtimes(outDir, legacyFilesList, created)
}

func chtimes(dir string, fileList []string, created time.Time) error {
	for _, fname := range fileList {
		if err := os.Chtimes(filepath.Join(dir, fname), created, created); err != nil {
			return err
		}
	}

	return nil
}

func gunzipLayer(dst, src, tag string, bar *progressbar.ProgressBar) error {
	gz := archive.NewGzip(dst, src)

	size, err := gz.GetUnarchSize()
	if err != nil {
		return err
	}
	bar.ContentLength(int64(size))
	bar.SetDescription(fmt.Sprintf("%s: %s ", tag, "Extracting"))
	bar.Flush()

	if _, err := gz.GunZip(bar); err != nil {
		return err
	}

	return os.Remove(src)
}
