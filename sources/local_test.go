// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sources

import (
	"testing"

	"github.com/keys-pub/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalUpdateSource(t *testing.T) {
	path := "../test/test.zip"
	jsonPath := "../test/update.json"
	local := NewLocalUpdateSource(path, jsonPath)
	assert.Equal(t, local.Description(), "Local")

	update, err := local.FindUpdate(updater.UpdateOptions{})
	require.NoError(t, err)
	require.NotNil(t, update)
}
