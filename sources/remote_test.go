// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keys-pub/updater"
	"github.com/keys-pub/updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteUpdateSource(t *testing.T) {
	jsonPath := "../test/update.json"
	data, err := util.ReadFile(jsonPath)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(data))
	}))

	local := NewRemoteUpdateSource(server.URL)
	assert.Equal(t, local.Description(), "Remote")

	update, err := local.FindUpdate(updater.UpdateOptions{})
	require.NoError(t, err)
	require.NotNil(t, update)
}
