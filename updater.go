// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

// Version is the updater version
const Version = "0.4.1"

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

func tempDir(appName string) string {
	return filepath.Join(os.TempDir(), "updater", appName)
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

func remove(tmpDir string) {
	if tmpDir != "" {
		logger.Infof("Clearing temporary directory: %q", tmpDir)
		if err := os.RemoveAll(tmpDir); err != nil {
			logger.Warningf("Error removing temporary directory %q: %s", tmpDir, err)
		}
	}
}
