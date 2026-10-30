package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestKeyGenerationAndPreservation(t *testing.T) {
	var output bytes.Buffer
	if err := run(strings.NewReader(`[{},{ }]`), &output); err != nil {
		t.Fatal(err)
	}
	var sets []keySet
	if err := json.Unmarshal(output.Bytes(), &sets); err != nil {
		t.Fatal(err)
	}
	for _, keys := range sets {
		public, _ := base64.RawURLEncoding.DecodeString(keys.PublicKey)
		private, _ := base64.RawURLEncoding.DecodeString(keys.PrivateKey)
		message := []byte("check generated signing pair")
		if !ed25519.Verify(public, message, ed25519.Sign(ed25519.PrivateKey(private), message)) {
			t.Fatal("generated keys cannot sign/verify")
		}
		restored, err := complete(keySet{PrivateKey: keys.PrivateKey, RefreshKey: keys.RefreshKey})
		if err != nil || restored != keys {
			t.Fatal("existing keys changed", err)
		}
	}
	before := append([]byte(nil), output.Bytes()...)
	output.Reset()
	if err := run(bytes.NewReader(before), &output); err != nil || !bytes.Equal(before, output.Bytes()) {
		t.Fatal("repeated init rotated keys", err)
	}
}

func TestRejectsBrokenOrSharedKeysWithoutOutput(t *testing.T) {
	keys, err := complete(keySet{})
	if err != nil {
		t.Fatal(err)
	}
	other, _ := complete(keySet{})
	cases := [][]keySet{
		{{PublicKey: keys.PublicKey}},
		{{PublicKey: other.PublicKey, PrivateKey: keys.PrivateKey}},
		{{PrivateKey: "bad"}},
		{{PrivateKey: keys.PrivateKey, RefreshKey: keys.PublicKey}},
		{keys, keys},
	}
	for _, sets := range cases {
		input, _ := json.Marshal(sets)
		var output bytes.Buffer
		if err := run(bytes.NewReader(input), &output); err == nil || output.Len() != 0 {
			t.Fatal("bad configuration produced secrets")
		}
	}
}

func TestRejectsCrossDomainRefreshKeyReuse(t *testing.T) {
	first, err := complete(keySet{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := complete(keySet{})
	if err != nil {
		t.Fatal(err)
	}
	private, _ := decode(second.PrivateKey, ed25519.PrivateKeySize)
	for _, reused := range []string{second.PublicKey, base64.RawURLEncoding.EncodeToString(private[:ed25519.SeedSize])} {
		first.RefreshKey = reused
		for _, sets := range [][]keySet{{first, second}, {second, first}} {
			input, _ := json.Marshal(sets)
			var output bytes.Buffer
			if err := run(bytes.NewReader(input), &output); err == nil || output.Len() != 0 {
				t.Fatal("cross-domain signing material accepted as a refresh recovery key")
			}
		}
	}
}
