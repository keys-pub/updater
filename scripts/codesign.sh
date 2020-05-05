#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

bin=$1

# TODO: Pass in .Os from goreleaser when that works
if [[ ! "$bin" = *"/updater_darwin_amd64/updater" ]]; then
    echo "Skipping unsupported platform"
    exit 0
fi


dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

code_sign_identity="Developer ID Application: Gabriel Handford (U2622K69A6)"
codesign --verbose --sign "$code_sign_identity" "$bin"
