package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ProtonMail/go-crypto/openpgp"
	"os"
	"strings"
)

func main() {
	// payload is the information read from the QR code. For simplicity, reading the QR code is skipped.
	// Reading the payload from a file is good enough for now.
	payload, err := os.ReadFile("payload.yaml")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read file: %v", err))
		return
	}

	// which public key to use to verify the signature should be included in the claims
	// while `i_cert` is defined for this, having URLs here is not the brightest idea
	err = verifySignature(string(payload), "public.asc")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to verify signature: %v", err))
		return
	}
}

// function to verify the signature
func verifySignature(payload string, pubKeyPath string) error {
	// parse payload
	payloadParts := strings.Split(payload, "\nx_sig: ")
	if len(payloadParts) != 2 {
		return errors.New("invalid payload format")
	}
	data := payloadParts[0]
	sig := payloadParts[1]

	// read public key
	publicKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read public key file: %v", err)
	}

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(publicKey))
	if err != nil {
		return fmt.Errorf("unable to parse PGP key: %v", err)
	}

	// hash the data
	sum := sha256.Sum256([]byte(data))
	hash := fmt.Sprintf("%x", sum)

	// verify the signature
	_, err = openpgp.CheckDetachedSignature(keyring, strings.NewReader(hash), base64.NewDecoder(base64.StdEncoding, strings.NewReader(sig)), nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}

	print(data)

	return nil
}
