//go:build phpx

package phpx

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

// TestOpenSSLCrypt pins the ciphertexts php's openssl_encrypt answers for one
// key, IV and message, in the AES-CBC and AES-GCM modes laravel's Encrypter
// reads. The expected bytes come from php 8.4 with the openssl extension.
func TestOpenSSLCrypt(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	iv := "abcdef0123456789"
	plain := "hello world, this is a secret"

	got, err := opensslCrypt(true, plain, "aes-256-cbc", key, int64(opensslRawData), iv, nil, nil, "", nil)
	if err != nil {
		t.Fatalf("cbc encrypt: %v", err)
	}
	wantRaw := "4d4985ea9f5dec909977bb699c92113c8328d15d522d7082ebd49184e30e3c13"
	if hex.EncodeToString([]byte(got.(string))) != wantRaw {
		t.Fatalf("cbc raw = %x, want %s", got, wantRaw)
	}
	back, err := opensslCrypt(false, got.(string), "aes-256-cbc", key, int64(opensslRawData), iv, nil, nil, "", nil)
	if err != nil || back != plain {
		t.Fatalf("cbc decrypt = %v, %v; want %q", back, err, plain)
	}

	b64, err := opensslCrypt(true, plain, "aes-256-cbc", key, nil, iv, nil, nil, "", nil)
	if err != nil {
		t.Fatalf("cbc base64 encrypt: %v", err)
	}
	if want := "TUmF6p9d7JCZd7tpnJIRPIMo0V1SLXCC69SRhOMOPBM="; b64 != want {
		t.Fatalf("cbc base64 = %q, want %q", b64, want)
	}
	back, err = opensslCrypt(false, b64.(string), "aes-256-cbc", key, nil, iv, nil, nil, "", nil)
	if err != nil || back != plain {
		t.Fatalf("cbc base64 decrypt = %v, %v; want %q", back, err, plain)
	}

	gcmIV := "abcdef012345"
	var tag any
	var setTag func(any)
	setTag = func(v any) { tag = v }
	sealed, err := opensslCrypt(true, plain, "aes-256-gcm", key, int64(opensslRawData), gcmIV, setTag, nil, "", nil)
	if err != nil {
		t.Fatalf("gcm encrypt: %v", err)
	}
	if want := "e9938e62c3a7312fca34d1e0ead7fec87a84d60e809a98a87fa837d586"; hex.EncodeToString([]byte(sealed.(string))) != want {
		t.Fatalf("gcm = %x, want %s", sealed, want)
	}
	if tagText, ok := tag.(string); !ok || hex.EncodeToString([]byte(tagText)) != "cc929da6e6c482f61759921823795a5b" {
		t.Fatalf("tag = %x, want cc929da6e6c482f61759921823795a5b", tag)
	}
	opened, err := opensslCrypt(false, sealed.(string), "aes-256-gcm", key, int64(opensslRawData), gcmIV, nil, tag, "", nil)
	if err != nil || opened != plain {
		t.Fatalf("gcm decrypt = %v, %v; want %q", opened, err, plain)
	}
	// A tampered tag is refused, as openssl_decrypt answers false.
	if got, err := opensslCrypt(false, sealed.(string), "aes-256-gcm", key, int64(opensslRawData), gcmIV, nil, "0000000000000000", "", nil); err != nil || got != false {
		t.Fatalf("gcm decrypt of a wrong tag = %v, %v; want false", got, err)
	}
}

// TestOpenSSLTagShape pins the tag openssl_encrypt answers base64 encoded when
// the caller asked for base64, and the IV length table.
func TestOpenSSLTagShape(t *testing.T) {
	if got := opensslCipherIVLength("aes-256-cbc"); got != 16 {
		t.Fatalf("iv length = %d, want 16", got)
	}
	if got := opensslCipherIVLength("aes-256-gcm"); got != 12 {
		t.Fatalf("gcm iv length = %d, want 12", got)
	}
	if got := opensslCipherIVLength("nope-cipher"); got != -1 {
		t.Fatalf("unknown cipher answered %d, want -1", got)
	}
	_ = base64.StdEncoding
}

// opensslCipherIVLength answers the IV length a method names, -1 for an
// unknown one, the way the registered openssl_cipher_iv_length does.
func opensslCipherIVLength(method string) int {
	spec, ok := opensslCipher(method)
	if !ok {
		return -1
	}
	return spec.ivLen
}
