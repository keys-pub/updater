// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// DigestType is type of digest.
type DigestType string

const (
	// SHA256 is sha256 digest type
	SHA256 DigestType = "sha256"
	// SHA512 is sha512 digest type
	SHA512 DigestType = "sha512"
)

// CheckDigest returns no error if digest matches file
func CheckDigest(digest string, path string, typ DigestType) error {
	if digest == "" {
		return fmt.Errorf("Missing digest")
	}
	calcDigest, err := DigestForFileAtPath(path, typ)
	if err != nil {
		return err
	}
	if calcDigest != digest {
		return fmt.Errorf("Invalid digest: %s != %s (%s)", calcDigest, digest, path)
	}
	logger.Infof("Verified digest: %s (%s)", digest, path)
	return nil
}

// DigestForFileAtPath returns a SHA256 digest for file at specified path
func DigestForFileAtPath(path string, typ DigestType) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer Close(f)

	switch typ {
	case SHA256, "":
		return Digest256(f)
	case SHA512:
		return Digest512(f)
	default:
		return "", errors.Errorf("invalid digest type: %s", typ)
	}
}

// Digest256 returns a SHA256 digest.
func Digest256(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	digest := hex.EncodeToString(hasher.Sum(nil))
	return digest, nil
}

// Digest512 returns a SHA256 digest.
func Digest512(r io.Reader) (string, error) {
	hasher := sha512.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	digest := hex.EncodeToString(hasher.Sum(nil))
	return digest, nil
}
