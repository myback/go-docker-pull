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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
)

type WWWAuthenticate struct {
	Realm, Service, Scope string
}

func WWWAuthenticateParse(s string) (out WWWAuthenticate) {
	headerParts := strings.SplitN(s, " ", 2)

	outs := reflect.ValueOf(&out).Elem()
	for _, part := range strings.Split(headerParts[1], ",") {
		kv := strings.SplitN(part, "=", 2)
		outs.FieldByName(strings.Title(kv[0])).SetString(strings.ReplaceAll(kv[1], "\"", ""))
	}

	return out
}

func (www *WWWAuthenticate) Url(action string) (string, error) {
	u, err := url.Parse(www.Realm)
	if err != nil {
		return "", nil
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", nil
	}

	q.Set("service", www.Service)
	if www.Scope != "" {
		if action != "" {
			scope := strings.Split(www.Scope, ":")
			if len(scope) == 3 {
				scope[3] = action
				www.Scope = strings.Join(scope, ":")
			}
		}

		q.Set("scope", www.Scope)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func FileHashEqual(filename, hash string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	return hex.EncodeToString(hasher.Sum(nil)) == hash, nil
}

func SaveToJson(file string, v interface{}) error {
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	return json.NewEncoder(fd).Encode(v)
}
