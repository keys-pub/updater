package main

import (
	"fmt"
	"path/filepath"

	"github.com/keys-pub/updater"
	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

type ctx struct {
	appName string
	version string
}

func newContext(appName string, version string) updater.Context {
	return &ctx{
		appName: appName,
		version: version,
	}
}

func (c *ctx) GetUpdateUI() updater.UpdateUI {
	return c
}

func (c *ctx) UpdatePrompt(update *updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	return &updater.UpdatePromptResponse{Action: updater.UpdateActionContinue}, nil
}

func (c *ctx) UpdateOptions() updater.UpdateOptions {
	return updater.UpdateOptions{
		AppName: c.appName,
		Version: c.version,
	}
}

func (c *ctx) Verify(update *updater.Update) error {
	// No extra verification by default (other than digest checks when downloading).
	if update.Asset.Signature != "" {
		return errors.Errorf("saltpack verification not supported")
	}
	return nil
}

func (c *ctx) BeforeUpdatePrompt(update *updater.Update, options updater.UpdateOptions) error {
	return nil
}

func (c *ctx) BeforeApply(update *updater.Update) error {
	return nil
}

func (c *ctx) check(sourcePath string, destinationPath string) error {
	// Check to make sure the update source path is a real directory
	ok, err := util.IsDirReal(sourcePath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Source path isn't a directory")
	}
	return nil
}

func (c *ctx) Apply(update *updater.Update, options updater.UpdateOptions, tmpDir string) error {
	localPath := update.Asset.LocalPath
	destinationPath := options.DestinationPath
	// The file name we unzip over should match the (base) file in the destination path
	filename := filepath.Base(destinationPath)
	return util.UnzipOver(localPath, filename, destinationPath, c.check, tmpDir)
}

func (c *ctx) AfterApply(update *updater.Update) error {
	return nil
}

func (c *ctx) ReportError(err error, update *updater.Update, options updater.UpdateOptions) {

}

func (c *ctx) ReportAction(updatePromptResponse updater.UpdatePromptResponse, update *updater.Update, options updater.UpdateOptions) {

}

func (c *ctx) ReportSuccess(update *updater.Update, options updater.UpdateOptions) {

}

func (c *ctx) AfterUpdateCheck(update *updater.Update) {

}
