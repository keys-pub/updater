package main

import (
	"os/exec"

	"github.com/keys-pub/updater"
	"github.com/pkg/errors"
)

func apply(options updater.UpdateOptions, assetPath string, applyPath string) error {
	logger.Infof("Running msiexec.exe -i %s", assetPath)
	cmd := exec.Command("msiexec.exe", "-i", assetPath)
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "Command failed (%s)", assetPath)
	}

	// TODO: Cleanup asset
	// logger.Infof("Removing %s", assetPath)
	// if err := os.Remove(assetPath); err != nil {
	// 	return err
	// }

	return nil
}
