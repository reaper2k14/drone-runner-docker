// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/drone-runners/drone-runner-docker/engine"
)

// config represents the Docker client configuration,
// typically located at ~/.docker/config.json
type config struct {
	Auths map[string]auths `json:"auths"`
}

type auths struct {
	Auth string `json:"auth"`
}

// Parse parses the registry credential from the reader.
func Parse(r io.Reader) ([]*engine.Auth, error) {
	c := new(config)
	err := json.NewDecoder(r).Decode(c)
	if err != nil {
		return nil, err
	}
	var auths []*engine.Auth
	for k, v := range c.Auths {
		username, password := decode(v.Auth)
		auths = append(auths, &engine.Auth{
			Address:  hostname(k),
			Username: username,
			Password: password,
		})
	}
	return auths, nil
}

// ParseFile parses the registry credential file.
func ParseFile(filepath string) ([]*engine.Auth, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseString parses the registry credential file.
func ParseString(s string) ([]*engine.Auth, error) {
	return Parse(strings.NewReader(s))
}

// encode returns the encoded credentials.
func encode(username, password string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(username + ":" + password),
	)
}

// decode returns the decoded credentials.
func decode(s string) (username, password string) {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return
	}
	parts := strings.SplitN(string(d), ":", 2)
	if len(parts) > 0 {
		username = parts[0]
	}
	if len(parts) > 1 {
		password = parts[1]
	}
	return
}

func hostname(s string) string {
	uri, _ := url.Parse(s)
	if uri.Host != "" {
		s = uri.Host
	}
	return s
}

// Encode returns the json marshaled, base64 encoded
// credential string that can be passed to the docker
// registry authentication header.
func Encode(username, password string) string {
	v := struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}{
		Username: username,
		Password: password,
	}
	buf, _ := json.Marshal(&v)
	return base64.URLEncoding.EncodeToString(buf)
}

// Marshal marshals the Auth credentials to a
// .docker/config.json file.
func Marshal(list []*engine.Auth) ([]byte, error) {
	out := &config{}
	out.Auths = map[string]auths{}
	for _, item := range list {
		out.Auths[item.Address] = auths{
			Auth: encode(
				item.Username,
				item.Password,
			),
		}
	}
	return json.Marshal(out)
}