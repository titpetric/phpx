//go:build phpx

package phpx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// The openssl_* calls a framework's encrypter makes: laravel's Encrypter asks
// for a cipher's IV length, encrypts a payload with AES-CBC or AES-GCM and
// decrypts it again, and reads the authentication tag the GCM modes answer
// through the by-reference parameter. The binding answers those ciphers with
// go's crypto packages; any other method is refused as php refuses an algorithm
// its openssl does not know.

func init() {
	runner.RegisterArea(`Runtime\PX\Openssl`, registerPhpxOpenSSL)
}

// php's openssl_encrypt options, whose low bits mean what they say.
const (
	opensslRawData     = 1
	opensslZeroPadding = 2
)

// registerPhpxOpenSSL installs the cipher functions and their constants.
func registerPhpxOpenSSL(rt *runner.Runtime) {
	rt.SetConst("OPENSSL_RAW_DATA", int64(opensslRawData))
	rt.SetConst("OPENSSL_ZERO_PADDING", int64(opensslZeroPadding))
	rt.SetConst("OPENSSL_DONT_ZERO_PAD_KEY", int64(4))
	rt.SetConst("OPENSSL_ALGO_SHA1", int64(1))
	rt.SetConst("OPENSSL_ALGO_MD5", int64(2))

	// openssl_cipher_iv_length answers the IV size in bytes of a cipher name,
	// and false for one the binding does not know.
	rt.RegisterFunc("openssl_cipher_iv_length", func(method string) any {
		spec, ok := opensslCipher(method)
		if !ok {
			return false
		}
		return int64(spec.ivLen)
	})
	// openssl_get_cipher_methods answers the cipher names this binding opens.
	rt.RegisterFunc("openssl_get_cipher_methods", func(...any) *model.Array {
		out := model.NewArray()
		for _, name := range []string{
			"aes-128-cbc", "aes-192-cbc", "aes-256-cbc",
			"aes-128-ctr", "aes-192-ctr", "aes-256-ctr",
			"aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
			"aes-128-ecb", "aes-192-ecb", "aes-256-ecb",
			"des-ede3-cbc",
		} {
			out.Append(name)
		}
		return out
	})
	// openssl_encrypt encrypts $data with $cipher_algo under $passphrase, with
	// $iv for the modes that take one, and reports the GCM tag through &$tag.
	// The answer is base64 unless OPENSSL_RAW_DATA is in $options.
	rt.RegisterFunc("openssl_encrypt", func(data, method, passphrase string, options any, iv string, tag func(any), tagValue any, aad string, tagLength any) (any, error) {
		return opensslCrypt(true, data, method, passphrase, options, iv, tag, tagValue, aad, tagLength)
	})
	// openssl_decrypt reverses openssl_encrypt, verifying the GCM tag it is given.
	rt.RegisterFunc("openssl_decrypt", func(data, method, passphrase string, options any, iv string, tag func(any), tagValue any, aad string, tagLength any) (any, error) {
		return opensslCrypt(false, data, method, passphrase, options, iv, tag, tagValue, aad, tagLength)
	})
	// openssl_random_pseudo_bytes answers $length cryptographically secure
	// bytes and reports through &$strong that the source is strong.
	rt.RegisterFunc("openssl_random_pseudo_bytes", func(length any, strong func(any)) (any, error) {
		n := int(phpval.Int(length))
		if n < 1 {
			return false, nil
		}
		buf := make([]byte, n)
		if _, err := rand.Read(buf); err != nil {
			if strong != nil {
				strong(false)
			}
			return false, nil
		}
		if strong != nil {
			strong(true)
		}
		return string(buf), nil
	})
	// openssl_error_string answers the last error the binding raised; it keeps
	// none, so there is never one.
	rt.RegisterFunc("openssl_error_string", func() any { return false })
}

// opensslSpec is one cipher the binding opens: the key and block sizes and
// what the mode does with the block.
type opensslSpec struct {
	keyLen int
	ivLen  int
	mode   string
}

// opensslCipher resolves a cipher name the way php matches one, ignoring case.
func opensslCipher(name string) (opensslSpec, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "aes-")
	parts := strings.Split(name, "-")
	if len(parts) == 2 {
		bits := map[string]int{"128": 16, "192": 24, "256": 32}[parts[0]]
		mode := parts[1]
		switch mode {
		case "cbc", "ctr", "ecb":
			return opensslSpec{keyLen: bits, ivLen: 16, mode: mode}, bits != 0
		case "gcm":
			return opensslSpec{keyLen: bits, ivLen: 12, mode: mode}, bits != 0
		}
	}
	if strings.EqualFold(name, "des-ede3-cbc") {
		return opensslSpec{keyLen: 24, ivLen: 8, mode: "cbc"}, true
	}
	return opensslSpec{}, false
}

// opensslCrypt runs one direction of openssl_encrypt and openssl_decrypt.
func opensslCrypt(encrypt bool, data, method, passphrase string, options any, iv string, tag func(any), tagValue any, aad string, tagLength any) (any, error) {
	spec, ok := opensslCipher(method)
	if !ok {
		return false, nil
	}
	opts := 0
	if options != nil {
		opts = int(phpval.Int(options))
	}
	key := opensslKey(passphrase, spec.keyLen)
	blockCipher, err := opensslBlock(strings.ToLower(strings.TrimSpace(method)), key)
	if err != nil {
		return false, nil
	}
	if len(iv) != spec.ivLen && spec.mode != "ecb" {
		return false, nil
	}
	if encrypt {
		dst, outTag, err := opensslEncrypt(spec, blockCipher, []byte(data), []byte(iv), []byte(aad), opts, tagLength)
		if err != nil {
			return false, nil
		}
		if outTag != nil && tag != nil {
			if opts&opensslRawData != 0 {
				tag(string(outTag))
			} else {
				tag(base64.StdEncoding.EncodeToString(outTag))
			}
		}
		if opts&opensslRawData != 0 {
			return string(dst), nil
		}
		return base64.StdEncoding.EncodeToString(dst), nil
	}
	raw := []byte(data)
	if opts&opensslRawData == 0 {
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return false, nil
		}
		raw = decoded
	}
	tagBytes, _ := tagValue.(string)
	plain, err := opensslDecrypt(spec, blockCipher, raw, []byte(iv), []byte(aad), []byte(tagBytes), opts, tagLength)
	if err != nil {
		return false, nil
	}
	return string(plain), nil
}

// opensslKey answers the key openssl uses: the cipher's length, truncated or
// zero padded.
func opensslKey(passphrase string, size int) []byte {
	key := make([]byte, size)
	copy(key, passphrase)
	return key
}

// opensslBlock builds the block cipher behind a method name.
func opensslBlock(method string, key []byte) (cipher.Block, error) {
	base := method
	for _, mode := range []string{"-cbc", "-ctr", "-gcm", "-ecb"} {
		if strings.HasSuffix(base, mode) {
			base = strings.TrimSuffix(base, mode)
			break
		}
	}
	switch {
	case strings.HasPrefix(base, "aes-"):
		return aes.NewCipher(key)
	case base == "des-ede3":
		return des.NewTripleDESCipher(key)
	}
	return nil, fmt.Errorf("unknown cipher %q", method)
}

// opensslEncrypt runs the mode's encryption, padding as php pads.
func opensslEncrypt(spec opensslSpec, block cipher.Block, plain, iv, aad []byte, opts int, tagLength any) ([]byte, []byte, error) {
	switch spec.mode {
	case "cbc":
		padded := plain
		if opts&opensslZeroPadding == 0 {
			padded = pkcs7Pad(plain, block.BlockSize())
		}
		out := make([]byte, len(padded))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
		return out, nil, nil
	case "ecb":
		padded := plain
		if opts&opensslZeroPadding == 0 {
			padded = pkcs7Pad(plain, block.BlockSize())
		}
		out := make([]byte, len(padded))
		size := block.BlockSize()
		for i := 0; i < len(padded); i += size {
			block.Encrypt(out[i:i+size], padded[i:i+size])
		}
		return out, nil, nil
	case "ctr":
		out := make([]byte, len(plain))
		cipher.NewCTR(block, iv).XORKeyStream(out, plain)
		return out, nil, nil
	case "gcm":
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, nil, err
		}
		tagLen := gcm.Overhead()
		if tagLength != nil {
			if n := int(phpval.Int(tagLength)); n > 0 && n <= gcm.Overhead() {
				gcm, err = cipher.NewGCMWithTagSize(block, n)
				if err != nil {
					return nil, nil, err
				}
				tagLen = n
			}
		}
		sealed := gcm.Seal(nil, iv, plain, aad)
		cut := len(sealed) - tagLen
		return sealed[:cut], sealed[cut:], nil
	}
	return nil, nil, fmt.Errorf("unsupported mode %q", spec.mode)
}

// opensslDecrypt runs the mode's decryption and strips the padding.
func opensslDecrypt(spec opensslSpec, block cipher.Block, raw, iv, aad, tag []byte, opts int, tagLength any) ([]byte, error) {
	switch spec.mode {
	case "cbc":
		if len(raw) == 0 || len(raw)%block.BlockSize() != 0 {
			return nil, fmt.Errorf("wrong final block length")
		}
		out := make([]byte, len(raw))
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, raw)
		return opensslUnpad(out, block.BlockSize(), opts), nil
	case "ecb":
		if len(raw) == 0 || len(raw)%block.BlockSize() != 0 {
			return nil, fmt.Errorf("wrong final block length")
		}
		out := make([]byte, len(raw))
		size := block.BlockSize()
		for i := 0; i < len(raw); i += size {
			block.Decrypt(out[i:i+size], raw[i:i+size])
		}
		return opensslUnpad(out, size, opts), nil
	case "ctr":
		out := make([]byte, len(raw))
		cipher.NewCTR(block, iv).XORKeyStream(out, raw)
		return out, nil
	case "gcm":
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		if tagLength != nil {
			if n := int(phpval.Int(tagLength)); n > 0 && n <= gcm.Overhead() {
				gcm, err = cipher.NewGCMWithTagSize(block, n)
				if err != nil {
					return nil, err
				}
			}
		}
		if len(tag) != gcm.Overhead() {
			return nil, fmt.Errorf("wrong tag length")
		}
		sealed := make([]byte, 0, len(raw)+len(tag))
		sealed = append(sealed, raw...)
		sealed = append(sealed, tag...)
		return gcm.Open(nil, iv, sealed, aad)
	}
	return nil, fmt.Errorf("unsupported mode %q", spec.mode)
}

// pkcs7Pad pads a message to the block size, always adding a full block when
// it already fits.
func pkcs7Pad(data []byte, size int) []byte {
	n := size - len(data)%size
	out := make([]byte, len(data)+n)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(n)
	}
	return out
}

// opensslUnpad strips the pkcs7 padding, and leaves the bytes alone when the
// caller asked for no padding.
func opensslUnpad(data []byte, size int, opts int) []byte {
	if opts&opensslZeroPadding != 0 || len(data) == 0 {
		return data
	}
	n := int(data[len(data)-1])
	if n <= 0 || n > size || n > len(data) {
		return data
	}
	return data[:len(data)-n]
}
