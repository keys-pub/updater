// +build darwin

package main

import (
	"path/filepath"
	"testing"

	"github.com/keys-pub/updater"
	"github.com/stretchr/testify/require"
)

func TestCheckDestination(t *testing.T) {
	applyPath := "/Applications/Keys.app"
	dir, file := filepath.Split(applyPath)
	err := checkDestination(updater.UpdateOptions{}, dir, file)
	require.NoError(t, err)

	applyPath = "/Applications"
	dir, file = filepath.Split(applyPath)
	err = checkDestination(updater.UpdateOptions{}, dir, file)
	require.EqualError(t, err, "invalid destination file: Applications")
}
