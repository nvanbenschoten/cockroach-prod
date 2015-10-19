// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Marc Berhault (marc@cockroachlabs.com)

// This is a slight modification of: https://github.com/docker/machine/blob/master/drivers/google/auth_util.go
// The main difference is that we have a single path for tokens, whereas docker-machine
// has --google-auth-token and a default store-path.
// Original license follows:

// Copyright 2014 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.

package google

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/cockroachdb/cockroach/util/log"
)

// OAuth logic. This initializes a GCE Service with a OAuth token.
// If the token (in Gob format) exists at 'authTokenPath', load it.
// Otherwise, redirect to the Google consent screen to get a code,
// generate a token from it, and save it in 'authTokenPath'.
//
// The token file format must be the same as that used by docker-machine.
const (
	authURL  = "https://accounts.google.com/o/oauth2/auth"
	tokenURL = "https://accounts.google.com/o/oauth2/token"
	// Cockroach client ID and secret.
	// TODO(marc): details show my personal email for now. We should have a more
	// generic user-facing one.
	clientID     = "962032490974-5avmqm15uklkgus98c7f862dk23u5mdk.apps.googleusercontent.com"
	clientSecret = "SSytmGLypTUPnj6a3PeV8LiR"
	redirectURI  = "urn:ietf:wg:oauth:2.0:oob"
)

var oauth2Config = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	Endpoint: oauth2.Endpoint{
		AuthURL:  authURL,
		TokenURL: tokenURL,
	},
	RedirectURL: redirectURI,
	Scopes:      []string{compute.ComputeScope},
}

var _ oauth2.TokenSource = browserSource{}
var _ oauth2.TokenSource = gobSource{}

// browserSource is a token source that punts to a browser for oauth.
type browserSource struct{}

func (source browserSource) Token() (*oauth2.Token, error) {
	// Get a new token. Pops up a browser window (hopefully).
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	authURL := oauth2Config.AuthCodeURL(randState)
	log.Infof("Opening auth URL in browser: %s", authURL)
	log.Infof("If the URL doesn't open please open it manually and copy the code here.")
	openURL(authURL)
	code := getCodeFromStdin()
	return oauth2Config.Exchange(context.Background(), code)
}

// gobSource is a gob-encoding file-backed token source.
type gobSource struct {
	path string
	base oauth2.TokenSource
}

// Token returns the cached token value, or an error if none is found.
func (f gobSource) Token() (*oauth2.Token, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return f.cacheUnderlying()
	}
	token := &oauth2.Token{}
	if err := gob.NewDecoder(file).Decode(token); err != nil {
		return f.cacheUnderlying()
	}

	return token, file.Close()
}

// TODO(marc): we should write to a tmp file and rename in case we error out.
func (f gobSource) cacheUnderlying() (token *oauth2.Token, err error) {
	token, err = f.base.Token()
	if err != nil {
		return token, err
	}

	// Create the parent directory if necessary.
	if mkdirErr := os.MkdirAll(filepath.Dir(f.path), 0700); mkdirErr != nil {
		return
	}

	file, openFileErr := os.OpenFile(f.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if openFileErr != nil {
		return
	}
	_ = gob.NewEncoder(file).Encode(token)
	_ = file.Close()

	return
}

func newOauthClient(authTokenPath string) (*http.Client, error) {
	token, err := oauth2.ReuseTokenSource(nil, gobSource{path: authTokenPath, base: browserSource{}}).Token()
	if err != nil {
		log.Infof("problem exchanging code: %s", err)
		return nil, err
	}
	return oauth2Config.Client(context.Background(), token), nil
}

func getCodeFromStdin() string {
	fmt.Print("Enter code: ")
	var code string
	fmt.Scanln(&code)
	return strings.Trim(code, "\n")
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		if err := exec.Command(bin, url).Run(); err == nil {
			return
		}
	}
}
