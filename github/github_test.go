package github

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/keys-pub/updater"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	s := newGithubSource("keys-pub/app")
	b, err := ioutil.ReadFile("./testdata/latest-mac.yml")
	require.NoError(t, err)
	upd, err := s.updateFromGithub(b, updater.UpdateOptions{Version: "0.0.17"})
	require.NoError(t, err)
	require.Equal(t, "0.0.18", upd.Version)
	require.Equal(t, int64(1583275443689), upd.PublishedAt)
	require.NotNil(t, upd.Asset)
	require.Equal(t, "Keys-0.0.18-mac.zip", upd.Asset.Name)
	require.Equal(t, "https://github.com/keys-pub/app/releases/latest/download/Keys-0.0.18-mac.zip", upd.Asset.URL)
	require.Equal(t, "9fe462603acbd84e55e5dfa6a02f40d0483551c88bd053b4b3827aba67d7fe3e53414a2214f6387a02e0bfc667d464ed0cc494f14b6ca04ae5ca81a20d503618", upd.Asset.Digest)
	require.True(t, upd.NeedUpdate)

	upd, err = s.updateFromGithub(b, updater.UpdateOptions{Version: "0.0.18"})
	require.NoError(t, err)
	require.False(t, upd.NeedUpdate)

	upd, err = s.updateFromGithub(b, updater.UpdateOptions{Version: "0.0.19"})
	require.NoError(t, err)
	require.False(t, upd.NeedUpdate)
}

func TestPrerelease(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	s := newGithubSource("keys-pub/app")
	urs, err := s.findManifestURL(true, time.Second*10)
	require.NoError(t, err)
	t.Logf("Prelease: %s", urs)

	urs2, err := s.findManifestURL(false, time.Second*10)
	require.NoError(t, err)
	t.Logf("Latest: %s", urs2)
	require.True(t, strings.HasPrefix(urs2, "https://github.com/keys-pub/app/releases/latest/download/latest"))
}
