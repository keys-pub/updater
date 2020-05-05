// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package github

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
	return fmt.Sprintf("github.com/%s", s.repo)
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

	url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", s.repo, gupd.Version, gupd.Path)

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

func (s githubSource) latestManifestURL() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("https://github.com/%s/releases/latest/download/latest-mac.yml", s.repo), nil
	case "windows":
		return fmt.Sprintf("https://github.com/%s/releases/latest/download/latest-windows.yml", s.repo), nil
	default:
		return "", errors.Errorf("Unsupported platform")
	}
}

func (s githubSource) tagManifestURL(tag string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("https://github.com/%s/releases/download/%s/latest-mac.yml", s.repo, tag), nil
	case "windows":
		return fmt.Sprintf("https://github.com/%s/releases/download/%s/latest-windows.yml", s.repo, tag), nil
	default:
		return "", errors.Errorf("Unsupported platform")
	}
}

type release struct {
	Prerelease bool   `json:"prerelease"`
	Name       string `json:"name"`
	Tag        string `json:"tag_name"`
}

func (s githubSource) prereleaseURL(timeout time.Duration) (string, error) {
	b, err := request(fmt.Sprintf("https://api.github.com/repos/%s/releases", s.repo), timeout)
	if err != nil {
		return "", err
	}

	var rels []*release
	if err := json.Unmarshal(b, &rels); err != nil {
		return "", err
	}
	if len(rels) == 0 {
		return "", nil
	}

	rel := rels[0]

	if !rel.Prerelease {
		return "", nil
	}

	if rel.Tag == "" {
		return "", errors.Errorf("no tag for release")
	}

	return s.tagManifestURL(rel.Tag)
}

func (s githubSource) findManifestURL(prerelease bool, timeout time.Duration) (string, error) {
	if prerelease {
		urs, err := s.prereleaseURL(timeout)
		// If prelease not found or errored, fall back to latest
		if err != nil {
			logger.Infof("Error checking for prerelease: %v", err)
		} else if urs != "" {
			return urs, nil
		}
	}

	return s.latestManifestURL()
}

func (s githubSource) findUpdate(options updater.UpdateOptions, timeout time.Duration) (*updater.Update, error) {
	if s.repo == "" {
		return nil, errors.Errorf("No repo specified")
	}

	manifestURL, err := s.findManifestURL(options.Prerelease, timeout)
	if err != nil {
		return nil, err
	}

	if manifestURL == "" {
		return nil, nil
	}

	ur, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}
	urs := ur.String()
	logger.Infof("Requesting %s", urs)

	b, err := request(urs, timeout)
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

func request(urs string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequest("GET", urs, nil)
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

	return ioutil.ReadAll(resp.Body)
}
