// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

// Version is the updater version
const Version = "0.4.1"

// Updater knows how to find and apply updates
type Updater struct {
	source       UpdateSource
	config       Config
	tickDuration time.Duration
}

// UpdateSource defines where the updater can find updates
type UpdateSource interface {
	// Description is a short description about the update source
	Description() string
	// FindUpdate finds an update given options
	FindUpdate(options UpdateOptions) (*Update, error)
}

// Context defines options, UI and hooks for the updater.
// This is where you can define custom behavior specific to your apps.
type Context interface {
	GetUpdateUI() UpdateUI
	UpdateOptions() UpdateOptions
	Verify(update *Update) error
	BeforeUpdatePrompt(update *Update, options UpdateOptions) error
	BeforeApply(update *Update) error
	Apply(update *Update, options UpdateOptions, tmpDir string) error
	AfterApply(update *Update) error
	ReportError(err error, update *Update, options UpdateOptions)
	ReportAction(updatePromptResponse UpdatePromptResponse, update *Update, options UpdateOptions)
	ReportSuccess(update *Update, options UpdateOptions)
	AfterUpdateCheck(update *Update)
}

// Config defines configuration for the Updater
type Config interface {
	GetUpdateAuto() (bool, bool)
	SetUpdateAuto(b bool) error
	GetUpdateAutoOverride() bool
	SetUpdateAutoOverride(bool) error
	GetInstallID() string
	SetInstallID(installID string) error
	IsLastUpdateCheckTimeRecent(d time.Duration) bool
	SetLastUpdateCheckTime()
	SetLastAppliedVersion(string) error
	GetLastAppliedVersion() string
}

// NewUpdater constructs an Updater
func NewUpdater(source UpdateSource, config Config) *Updater {
	return &Updater{
		source:       source,
		config:       config,
		tickDuration: DefaultTickDuration,
	}
}

// SetTickDuration sets how often checks are.
func (u *Updater) SetTickDuration(dur time.Duration) {
	u.tickDuration = dur
}

// Update checks, downloads and performs an update
func (u *Updater) Update(ctx Context) (*Update, error) {
	options := ctx.UpdateOptions()
	update, err := u.update(ctx, options)
	report(ctx, err, update, options)
	return update, err
}

// Download an update.
// If downloaded update.Asset.LocalPath is set to downloaded path.
func (u *Updater) Download(ctx Context, update *Update, options UpdateOptions) error {
	// Linux updates don't have assets so it's ok to prompt for update above before
	// we check for nil asset.
	if update.Asset == nil || update.Asset.URL == "" {
		logger.Infof("No update asset to apply")
		return nil
	}

	tmpDir := tempDir(options.AppName)
	if err := u.downloadAsset(update.Asset, tmpDir, options); err != nil {
		return downloadErr(err)
	}

	logger.Infof("Verify asset: %s", update.Asset.LocalPath)
	if err := ctx.Verify(update); err != nil {
		return verifyErr(err)
	}

	return nil
}

func tempDir(appName string) string {
	return filepath.Join(os.TempDir(), "updater", appName)
}

// update returns the update received, and an error if the update was not
// performed. The error with be of type Error. The error may be due to the user
// (or system) canceling an update, in which case error.IsCancel() will be true.
func (u *Updater) update(ctx Context, options UpdateOptions) (*Update, error) {
	update, err := u.checkForUpdate(ctx, options)
	if err != nil {
		return nil, findErr(err)
	}
	if update == nil || !update.NeedUpdate {
		// No update available
		return nil, nil
	}
	logger.Infof("Got update with version: %s", update.Version)

	err = ctx.BeforeUpdatePrompt(update, options)
	if err != nil {
		return update, err
	}

	// Prompt for update
	updatePromptResponse, err := u.promptForUpdateAction(ctx, update, options)
	if err != nil {
		return update, promptErr(err)
	}
	switch updatePromptResponse.Action {
	case UpdateActionApply:
		ctx.ReportAction(updatePromptResponse, update, options)
	case UpdateActionAuto:
		ctx.ReportAction(updatePromptResponse, update, options)
	case UpdateActionSnooze:
		ctx.ReportAction(updatePromptResponse, update, options)
		return update, CancelErr(fmt.Errorf("Snoozed update"))
	case UpdateActionCancel:
		ctx.ReportAction(updatePromptResponse, update, options)
		return update, CancelErr(fmt.Errorf("Canceled"))
	case UpdateActionError:
		return update, promptErr(fmt.Errorf("Unknown prompt error"))
	case UpdateActionContinue:
		// Continue
	}

	// Linux updates don't have assets so it's ok to prompt for update above before
	// we check for nil asset.
	if update.Asset == nil || update.Asset.URL == "" {
		logger.Infof("No update asset to apply")
		return update, nil
	}

	tmpDir := tempDir(options.AppName)
	if err := u.downloadAsset(update.Asset, tmpDir, options); err != nil {
		return update, downloadErr(err)
	}

	logger.Infof("Verify asset: %s", update.Asset.LocalPath)
	if err := ctx.Verify(update); err != nil {
		return update, verifyErr(err)
	}

	if err := u.apply(ctx, update, options, tmpDir); err != nil {
		return update, err
	}

	return update, nil
}

func (u *Updater) apply(ctx Context, update *Update, options UpdateOptions, tmpDir string) error {
	logger.Infof("Before apply")
	if err := ctx.BeforeApply(update); err != nil {
		return applyErr(err)
	}

	logger.Infof("Applying update")
	if err := ctx.Apply(update, options, tmpDir); err != nil {
		logger.Infof("Apply error: %v", err)
		return applyErr(err)
	}

	logger.Infof("After apply")
	if err := ctx.AfterApply(update); err != nil {
		return applyErr(err)
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

// checkForUpdate checks a update source (like a remote API) for an update.
// It may set an InstallID, if the server tells us to.
func (u *Updater) checkForUpdate(ctx Context, options UpdateOptions) (*Update, error) {
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

	// Save InstallID if we received one
	if update.InstallID != "" && u.config.GetInstallID() != update.InstallID {
		logger.Debugf("Saving install ID: %s", update.InstallID)
		if err := u.config.SetInstallID(update.InstallID); err != nil {
			logger.Warningf("Error saving install ID: %s", err)
			ctx.ReportError(configErr(fmt.Errorf("Error saving install ID: %s", err)), update, options)
		}
	}

	return update, nil
}

// CheckForUpdate returns update.
func (u *Updater) CheckForUpdate(ctx Context) (*Update, error) {
	return u.checkForUpdate(ctx, ctx.UpdateOptions())
}

// promptForUpdateAction prompts the user for permission to apply an update
func (u *Updater) promptForUpdateAction(ctx Context, update *Update, options UpdateOptions) (UpdatePromptResponse, error) {
	logger.Debugf("Prompt for update")

	auto, autoSet := u.config.GetUpdateAuto()
	autoOverride := u.config.GetUpdateAutoOverride()
	logger.Debugf("Auto update: %s (set=%s autoOverride=%s)", strconv.FormatBool(auto), strconv.FormatBool(autoSet), strconv.FormatBool(autoOverride))
	if auto && !autoOverride {
		return UpdatePromptResponse{UpdateActionAuto, false, 0}, nil
	}

	updateUI := ctx.GetUpdateUI()

	// If auto update never set, default to true
	autoUpdate := auto || !autoSet
	promptOptions := UpdatePromptOptions{AutoUpdate: autoUpdate}
	updatePromptResponse, err := updateUI.UpdatePrompt(update, options, promptOptions)
	if err != nil {
		return UpdatePromptResponse{UpdateActionError, false, 0}, err
	}
	if updatePromptResponse == nil {
		return UpdatePromptResponse{UpdateActionError, false, 0}, fmt.Errorf("No response")
	}

	if updatePromptResponse.Action != UpdateActionContinue {
		logger.Debugf("Update prompt response: %#v", updatePromptResponse)
		if err := u.config.SetUpdateAuto(updatePromptResponse.AutoUpdate); err != nil {
			logger.Warningf("Error setting auto preference: %s", err)
			ctx.ReportError(configErr(fmt.Errorf("Error setting auto preference: %s", err)), update, options)
		}
	}

	return *updatePromptResponse, nil
}

func report(ctx Context, err error, update *Update, options UpdateOptions) {
	if err != nil {
		// Don't report cancels or GUI busy
		if e, ok := err.(Error); ok {
			if e.IsCancel() {
				return
			}
		}
		ctx.ReportError(err, update, options)
	} else if update != nil {
		ctx.ReportSuccess(update, options)
	}
}

func remove(tmpDir string) {
	if tmpDir != "" {
		logger.Infof("Clearing temporary directory: %q", tmpDir)
		if err := os.RemoveAll(tmpDir); err != nil {
			logger.Warningf("Error removing temporary directory %q: %s", tmpDir, err)
		}
	}
}
