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
	if file != options.AppName+".app" {
		return errors.Errorf("invalid destination file: %s", file)
	}
	return nil
}

func apply(options updater.UpdateOptions, sourcePath string, destinationPath string) error {
	destinationDir, destinationFile := filepath.Split(destinationPath)
	if err := checkDestination(options, destinationDir, destinationFile); err != nil {
		return err
	}

	sourceDir := path.Dir(sourcePath)
	args := []string{"/usr/bin/ditto", "-V", "-x", "-k", "--sequesterRsrc", "--rsrc", sourcePath, sourceDir}
	logger.Infof("Running %s", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Command failed (apply with ditto)")
	}
	logger.Debugf("%s", out)

	if _, err := os.Stat(destinationPath); err == nil {
		logger.Infof("Removing %s", destinationPath)
		if err := os.RemoveAll(destinationPath); err != nil {
			return err
		}
	}
	path := filepath.Join(sourceDir, destinationFile)
	logger.Infof("Renaming %s to %s", path, destinationPath)
	if err := os.Rename(path, destinationPath); err != nil {
		return err
	}

	logger.Infof("Removing %s", sourcePath)
	if err := os.Remove(sourcePath); err != nil {
		return err
	}

	return nil
}
