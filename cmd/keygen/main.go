// keygen completes and validates local Ed25519 key sets over JSON stdin/stdout.
// It uses only the standard library so make init also works in a Go container.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type keySet struct {
	PublicKey  string
	PrivateKey string
	RefreshKey string
}

func decode(value string, size int) ([]byte, error) {
	valueBytes, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(valueBytes) != size || base64.RawURLEncoding.EncodeToString(valueBytes) != value {
		return nil, errors.New("invalid key encoding or length")
	}
	return valueBytes, nil
}

func complete(keys keySet) (keySet, error) {
	var private ed25519.PrivateKey
	if keys.PrivateKey == "" {
		if keys.PublicKey != "" {
			return keys, errors.New("public key exists without its private key; refusing to replace it")
		}
		_, generated, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return keys, err
		}
		private = generated
		keys.PrivateKey = base64.RawURLEncoding.EncodeToString(private)
	} else {
		decoded, err := decode(keys.PrivateKey, ed25519.PrivateKeySize)
		if err != nil {
			return keys, errors.New("invalid Ed25519 private key")
		}
		private = ed25519.NewKeyFromSeed(decoded[:ed25519.SeedSize])
		if !bytes.Equal(decoded, private) {
			return keys, errors.New("inconsistent Ed25519 private key")
		}
	}
	public := base64.RawURLEncoding.EncodeToString(private.Public().(ed25519.PublicKey))
	if keys.PublicKey != "" && keys.PublicKey != public {
		return keys, errors.New("Ed25519 public/private key mismatch")
	}
	keys.PublicKey = public
	if keys.RefreshKey == "" {
		refresh := make([]byte, 32)
		if _, err := rand.Read(refresh); err != nil {
			return keys, err
		}
		keys.RefreshKey = base64.RawURLEncoding.EncodeToString(refresh)
	}
	refresh, err := decode(keys.RefreshKey, 32)
	if err != nil || bytes.Equal(refresh, private.Seed()) || bytes.Equal(refresh, private.Public().(ed25519.PublicKey)) {
		return keys, errors.New("refresh recovery requires an independent 32-byte key")
	}
	return keys, nil
}

func run(input io.Reader, output io.Writer) error {
	var sets []keySet
	decoder := json.NewDecoder(io.LimitReader(input, 65536))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sets); err != nil {
		return errors.New("invalid key-set input")
	}
	if len(sets) == 0 || len(sets) > 16 {
		return errors.New("expected between 1 and 16 key sets")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("unexpected trailing input")
	}
	seenPublic, seenRefresh := map[string]bool{}, map[string]bool{}
	signingMaterial := map[string]bool{}
	for i := range sets {
		completed, err := complete(sets[i])
		if err != nil {
			return fmt.Errorf("key set %d: %w", i+1, err)
		}
		if seenPublic[completed.PublicKey] || seenRefresh[completed.RefreshKey] {
			return errors.New("identity domains must use independent keys")
		}
		seenPublic[completed.PublicKey], seenRefresh[completed.RefreshKey] = true, true
		private, _ := decode(completed.PrivateKey, ed25519.PrivateKeySize) // Validated by complete.
		signingMaterial[completed.PublicKey] = true
		signingMaterial[base64.RawURLEncoding.EncodeToString(private[:ed25519.SeedSize])] = true
		sets[i] = completed
	}
	for _, keys := range sets {
		if signingMaterial[keys.RefreshKey] {
			return errors.New("refresh recovery keys must be independent of every identity domain's signing keys")
		}
	}
	return json.NewEncoder(output).Encode(sets)
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
