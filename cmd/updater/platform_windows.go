package main

import (
	"os"
	"os/exec"

	"github.com/keys-pub/updater"
	"github.com/pkg/errors"
)

func apply(options updater.UpdateOptions, assetPath string, applyPath string) error {
	cmd := exec.Command(assetPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Command failed (%s)", assetPath)
	}
	logger.Debugf("%s", out)

	logger.Infof("Removing %s", assetPath)
	if err := os.Remove(assetPath); err != nil {
		return err
	}

	return nil
}
