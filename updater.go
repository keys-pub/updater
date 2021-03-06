// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

// Version is the updater version
const Version = "0.2.3"

// Updater knows how to find and apply updates
type Updater struct {
	source UpdateSource
}

// UpdateSource defines where the updater can find updates
type UpdateSource interface {
	// Description is a short description about the update source
	Description() string
	// FindUpdate finds an update given options
	FindUpdate(options UpdateOptions) (*Update, error)
}

// NewUpdater constructs an Updater
func NewUpdater(source UpdateSource) *Updater {
	return &Updater{
		source: source,
	}
}

// Download an update.
// If downloaded update.Asset.LocalPath is set to downloaded path.
func (u *Updater) Download(update *Update, options UpdateOptions) error {
	// Linux updates don't have assets so it's ok to prompt for update above before
	// we check for nil asset.
	if update.Asset == nil || update.Asset.URL == "" {
		logger.Infof("No update asset to apply")
		return nil
	}

	tmpDir := tempDir(options.AppName)
	if err := u.downloadAsset(update.Asset, tmpDir, options); err != nil {
		return err
	}

	return nil
}

// downloadAsset will download the update to a temporary path (if not cached),
// check the digest, and set the LocalPath property on the asset.
func (u *Updater) downloadAsset(asset *Asset, tmpDir string, options UpdateOptions) error {
	if asset == nil {
		return fmt.Errorf("No asset to download")
	}

	var digestType util.DigestType
	switch asset.DigestType {
	case "", "sha256":
		digestType = util.SHA256
	case "sha512":
		digestType = util.SHA512
	default:
		return errors.Errorf("Unsupported digest type: %s", asset.DigestType)
	}

	downloadOptions := util.DownloadURLOptions{
		Digest:     asset.Digest,
		DigestType: digestType,
		UseETag:    true,
	}

	downloadPath := filepath.Join(tmpDir, asset.Name)
	// If asset had a file extension, lets add it back on
	if err := util.DownloadURL(asset.URL, downloadPath, downloadOptions); err != nil {
		return err
	}

	asset.LocalPath = downloadPath
	return nil
}

// CheckForUpdate checks a update source for an update.
func (u *Updater) CheckForUpdate(options UpdateOptions) (*Update, error) {
	logger.Infof("Checking for update, current version is %s", options.Version)
	logger.Infof("Using updater source: %s", u.source.Description())
	logger.Debugf("Using options: %#v", options)

	update, findErr := u.source.FindUpdate(options)
	if findErr != nil {
		return nil, findErr
	}
	if update == nil {
		return nil, nil
	}

	return update, nil
}

func tempDir(appName string) string {
	return filepath.Join(os.TempDir(), "updater", appName)
}

// Cleanup files, except for a path (presumably the update).
// You can do this after you download an update, so that if the update already
// exists it doesn't have to be re-downloaded, which removes all other files
// except the current update.
func Cleanup(appName string, except string) {
	dir := tempDir(appName)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Errorf("Error listing temp dir: %v", err)
		return
	}
	exceptBase := filepath.Base(except)

	logger.Infof("Cleanup files (except=%s)...", exceptBase)
	for _, file := range files {
		name := file.Name()
		if name != exceptBase {
			logger.Infof("Cleanup, removing %s", name)
			if err := os.RemoveAll(name); err != nil {
				logger.Errorf("Error removing %s: %v", file, err)
				return
			}
		}
	}
}
