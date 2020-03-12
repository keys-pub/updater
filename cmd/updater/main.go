// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/keys-pub/updater"
	"github.com/keys-pub/updater/github"
	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

type flags struct {
	version   bool
	logToFile bool
	appName   string
	github    string
	current   string
	download  bool
	apply     string
}

func main() {
	f := loadFlags()
	if err := run(f); err != nil {
		logFatal(err)
	}
}

func loadFlags() flags {
	f := flags{}
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.BoolVar(&f.logToFile, "log-to-file", false, "Log to file")
	flag.StringVar(&f.appName, "app-name", "", "App name")
	flag.StringVar(&f.github, "github", "", "Github repo")
	flag.StringVar(&f.current, "current", "", "Current version")
	flag.BoolVar(&f.download, "download", false, "Download update")
	flag.StringVar(&f.apply, "apply", "", "Apply")
	flag.Parse()
	return f
}

func logFatal(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

func run(f flags) error {
	if f.version {
		fmt.Printf("%s\n", updater.Version)
		return nil
	}

	log := NewLogger(InfoLevel)
	SetLogger(log)
	updater.SetLogger(log)
	util.SetLogger(log)
	github.SetLogger(log)

	if f.current == "" {
		return errors.Errorf("No current version specified (-current)")
	}
	if f.appName == "" {
		return errors.Errorf("No app name specified (-app-name)")
	}

	options := updater.UpdateOptions{
		AppName: f.appName,
		Version: f.current,
	}

	var src updater.UpdateSource
	if f.github != "" {
		src = github.NewUpdateSource(f.github)
	} else {
		return errors.Errorf("No update source")
	}

	upd := updater.NewUpdater(src)

	update, err := upd.CheckForUpdate(options)
	if err != nil {
		return err
	}
	if update == nil {
		fmt.Println("{}")
		return nil
	}

	checkOnly := !f.download && f.apply == ""

	if checkOnly || !update.NeedUpdate {
		b, err := json.MarshalIndent(update, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	// Download
	if f.download {
		if err := upd.Download(update, options); err != nil {
			return err
		}
		if update.Asset != nil {
			updater.Cleanup(options.AppName, update.Asset.LocalPath)
		}
	}

	// Apply
	if f.apply != "" {
		localPath := update.Asset.LocalPath
		if localPath == "" {
			return errors.Errorf("No local asset to apply, use with -download option?")
		}
		if err := apply(options, localPath, f.apply); err != nil {
			return err
		}
		update.Applied = f.apply
	}

	b, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
