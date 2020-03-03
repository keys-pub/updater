// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package saltpack

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	sp "github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
	"github.com/keys-pub/updater/util"
)

// VerifyDetachedFileAtPath verifies a file
func VerifyDetachedFileAtPath(path string, signature string, validKIDs map[string]bool) error {
	file, err := os.Open(path)
	defer util.Close(file)
	if err != nil {
		return err
	}
	err = VerifyDetached(file, signature, validKIDs)
	if err != nil {
		return fmt.Errorf("Error verifying signature: %s", err)
	}
	return nil
}

func checkSender(key sp.BasePublicKey, validKIDs map[string]bool) error {
	if key == nil {
		return fmt.Errorf("No key")
	}
	kid := key.ToKID()
	if kid == nil {
		return fmt.Errorf("No KID for key")
	}
	skid := hex.EncodeToString(kid)
	logger.Infof("Signed by %s", skid)
	if !validKIDs[skid] {
		return fmt.Errorf("Unknown signer KID: %s", skid)
	}
	logger.Debugf("Valid KID: %s", skid)
	return nil
}

// VerifyDetached verifies a message signature
func VerifyDetached(reader io.Reader, signature string, validKIDs map[string]bool) error {
	if reader == nil {
		return fmt.Errorf("No reader")
	}
	check := func(key sp.BasePublicKey) error {
		return checkSender(key, validKIDs)
	}
	return VerifyDetachedCheckSender(reader, []byte(signature), check)
}

// VerifyDetachedCheckSender verifies a message signature
func VerifyDetachedCheckSender(message io.Reader, signature []byte, checkSender func(sp.BasePublicKey) error) error {
	kr := basic.NewKeyring()
	var skey sp.SigningPublicKey
	var err error
	skey, _, err = sp.Dearmor62VerifyDetachedReader(sp.CheckKnownMajorVersion, message, string(signature), kr)
	if err != nil {
		return err
	}

	if err = checkSender(skey); err != nil {
		return err
	}

	return nil
}
