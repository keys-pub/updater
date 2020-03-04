// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"io"
	"net/http"

	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testZipPath = "./test/test.zip"

func newTestUpdater(t *testing.T) (*Updater, error) {
	return newTestUpdaterWithServer(t, nil, nil)
}

func newTestUpdaterWithServer(t *testing.T, testServer *httptest.Server, update *Update) (*Updater, error) {
	return NewUpdater(testUpdateSource{testServer: testServer, update: update}), nil
}

type testUpdateSource struct {
	testServer *httptest.Server
	update     *Update
	findErr    error
}

func (u testUpdateSource) Description() string {
	return "Test"
}

func testUpdate(uri string) *Update {
	return newTestUpdate(uri, true)
}

func newTestUpdate(uri string, needUpdate bool) *Update {
	update := &Update{
		Version:    "1.0.1",
		NeedUpdate: needUpdate,
	}
	if uri != "" {
		update.Asset = &Asset{
			Name:       "test.zip",
			URL:        uri,
			Digest:     "54970995e4d02da631e0634162ef66e2663e0eee7d018e816ac48ed6f7811c84", // shasum -a 256 test/test.zip
			DigestType: "sha256",
		}
	}
	return update
}

func (u testUpdateSource) FindUpdate(options UpdateOptions) (*Update, error) {
	return u.update, u.findErr
}

func testUpdateOptions() UpdateOptions {
	return UpdateOptions{
		Version: "1.0.0",
		AppName: "Keys",
	}
}

func testServerForUpdateFile(t *testing.T, path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(path)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/zip")
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	}))
}

func testServerForError(t *testing.T, err error) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), 500)
	}))
}

func testServerNotFound(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", 404)
	}))
}

func TestUpdaterCheck(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL))
	assert.NoError(t, err)
	options := testUpdateOptions()
	update, err := upr.CheckForUpdate(options)
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Logf("Update: %#v\n", update)
	require.NotNil(t, update.Asset)
	t.Logf("Asset: %#v\n", update.Asset)
	// TODO: Test
}

func TestUpdaterDownloadError(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("bad response"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL))
	assert.NoError(t, err)
	options := testUpdateOptions()
	update, err := upr.CheckForUpdate(options)
	require.NoError(t, err)
	require.NotNil(t, update)
	err = upr.Download(update, options)
	assert.EqualError(t, err, "500 Internal Server Error")
	// TODO: Test
}
