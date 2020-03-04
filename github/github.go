// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package github

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/blang/semver"
	"github.com/keys-pub/updater"
	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type githubSource struct {
	repo string
}

type file struct {
	URL          string `yaml:"url"`
	SHA512       string `yaml:"sha512"`
	Size         int    `yaml:"size"`
	BlockMapSize int    `yaml:"blockMapSize"`
}

type update struct {
	Version     string `yaml:"version"`
	Path        string `yaml:"path"`
	SHA512      string `yaml:"sha512"`
	ReleaseDate string `yaml:"releaseDate"`
	Files       []file `yaml:"files"`
}

// NewUpdateSource returns Github update source.
func NewUpdateSource(repo string) updater.UpdateSource {
	return newGithubSource(repo)
}

func newGithubSource(repo string) githubSource {
	return githubSource{repo: repo}
}

func (s githubSource) Description() string {
	return fmt.Sprintf("github:%s", s.repo)
}

func (s githubSource) FindUpdate(options updater.UpdateOptions) (*updater.Update, error) {
	return s.findUpdate(options, time.Minute)
}

func base64ToHex(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s githubSource) updateFromGithub(b []byte, options updater.UpdateOptions) (*updater.Update, error) {
	var gupd update
	if err := yaml.Unmarshal(b, &gupd); err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, gupd.ReleaseDate)
	if err != nil {
		return nil, err
	}
	ts := util.TimeToMillis(t)

	digest, err := base64ToHex(gupd.SHA512)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", s.repo, gupd.Path)

	curr, err := semver.Make(options.Version)
	next, err := semver.Make(gupd.Version)
	needUpdate := curr.LT(next)

	uu := &updater.Update{
		Version:     gupd.Version,
		PublishedAt: int64(ts),
		Asset: &updater.Asset{
			Name:       gupd.Path,
			URL:        url,
			Digest:     digest,
			DigestType: "sha512",
			// Signature: "",
		},
		NeedUpdate: needUpdate,
	}
	return uu, nil
}

func (s githubSource) findUpdate(options updater.UpdateOptions, timeout time.Duration) (*updater.Update, error) {
	if s.repo == "" {
		return nil, errors.Errorf("No repo specified")
	}

	var us string
	switch runtime.GOOS {
	case "darwin":
		us = fmt.Sprintf("https://github.com/%s/releases/latest/download/latest-mac.yml", s.repo)
	default:
		return nil, errors.Errorf("Unsupported platform")
	}

	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}
	urlString := u.String()
	logger.Infof("Requesting %s", urlString)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	client := util.HTTPClient(timeout)

	resp, err := client.Do(req)
	defer util.DiscardAndCloseBodyIgnoreError(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Find update returned bad HTTP status %v", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	uu, err := s.updateFromGithub(b, options)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Received update response: %#v", uu)

	return uu, nil
}
