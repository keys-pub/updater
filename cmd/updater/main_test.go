// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	f := flags{current: "0.0.1", appName: "Keys", github: "keys-pub/app"}
	err := run(f)
	require.NoError(t, err)

	f = flags{current: "0.0.1", appName: "Keys", github: "keys-pub/app", download: true}
	err = run(f)
	require.NoError(t, err)
}
