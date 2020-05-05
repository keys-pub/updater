// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

// Asset describes a downloadable file
type Asset struct {
	// Name of file.
	Name string `json:"name"`
	// URL to request from.
	URL string `json:"url"`
	// Digest is hex encoded digest.
	Digest string `json:"digest"`
	// DigestType is sha256 by default. Also supports "sha512".
	DigestType string `json:"digestType"`
	// LocalPath is where downloaded file resides.
	LocalPath string `json:"localPath"`
}

// Property is a generic key value pair for custom properties
type Property struct {
	Name  string `codec:"name" json:"name"`
	Value string `codec:"value" json:"value"`
}

// Update describes an update.
// If update is needed, NeedUpdate will be true.
// If update is downloaded, Asset.LocalPath will be set.
// If update was applied, Applied is set to the destination.
type Update struct {
	Version     string     `json:"version"`
	PublishedAt int64      `json:"publishedAt"`
	Props       []Property `codec:"props" json:"props,omitempty"`
	Asset       *Asset     `json:"asset,omitempty"`
	NeedUpdate  bool       `json:"needUpdate"`
	Applied     string     `json:"applied"`
}

// UpdateOptions are options used to find an update
type UpdateOptions struct {
	// Version is the current version of the app
	Version string `json:"version"`
	// AppName is name of the app
	AppName string `json:"appName"`
	// Prerelease will request latest prerelease
	Prerelease bool `json:"prerelease"`
}
