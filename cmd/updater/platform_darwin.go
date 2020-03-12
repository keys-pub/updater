package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/keys-pub/updater"
	"github.com/pkg/errors"
)

func checkDestination(options updater.UpdateOptions, dir string, file string) error {
	if strings.HasPrefix(file, ".app") {
		return errors.Errorf("invalid destination file: %s", file)
	}
	return nil
}

func apply(options updater.UpdateOptions, assetPath string, applyPath string) error {
	destinationDir, destinationFile := filepath.Split(applyPath)
	if err := checkDestination(options, destinationDir, destinationFile); err != nil {
		return err
	}

	sourceDir := path.Dir(assetPath)
	args := []string{"/usr/bin/ditto", "-V", "-x", "-k", "--sequesterRsrc", "--rsrc", assetPath, sourceDir}
	logger.Infof("Running %s", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Command failed (apply with ditto)")
	}
	logger.Debugf("%s", out)

	if _, err := os.Stat(applyPath); err == nil {
		logger.Infof("Removing existing %s", applyPath)
		if err := os.RemoveAll(applyPath); err != nil {
			return err
		}
	}
	path := filepath.Join(sourceDir, destinationFile)
	logger.Infof("Moving %s to %s", path, applyPath)
	if err := os.Rename(path, applyPath); err != nil {
		return err
	}

	return nil
}
