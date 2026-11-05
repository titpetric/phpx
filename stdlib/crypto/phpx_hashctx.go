//go:build phpx

package crypto

import (
	"crypto/hmac"
	"hash"
	"os"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxHashContext)
}

// HashContext is the value hash_init returns: a digest that takes data in
// pieces. Final consumes it, as php's does.
type HashContext struct {
	h     hash.Hash
	algo  string
	key   string
	hmac  bool
	final bool
}

// registerPhpxHashContext installs the incremental hash family and hash_file.
func registerPhpxHashContext(rt *runner.Runtime) {
	rt.RegisterFunc("hash_init", func(algo string, rest ...any) (*HashContext, error) {
		factory, ok := digests[algo]
		if !ok {
			return nil, unknownAlgo("hash_init", algo)
		}
		ctx := &HashContext{algo: algo}
		if len(rest) > 0 {
			if flags, ok := rest[0].(int64); ok && flags&1 != 0 {
				if len(rest) < 2 {
					return nil, errHashKey
				}
				ctx.hmac = true
				ctx.key, _ = rest[1].(string)
			}
		}
		if ctx.hmac {
			ctx.h = hmac.New(factory, []byte(ctx.key))
		} else {
			ctx.h = factory()
		}
		return ctx, nil
	})
	rt.RegisterFunc("hash_update", func(ctx *HashContext, data string) (bool, error) {
		if ctx == nil || ctx.final {
			return false, errHashFinalized("hash_update")
		}
		ctx.h.Write([]byte(data))
		return true, nil
	})
	rt.RegisterFunc("hash_final", func(ctx *HashContext, binary ...bool) (string, error) {
		if ctx == nil || ctx.final {
			return "", errHashFinalized("hash_final")
		}
		ctx.final = true
		return digestString(ctx.h.Sum(nil), binary), nil
	})
	rt.RegisterFunc("hash_copy", func(ctx *HashContext) (*HashContext, error) {
		if ctx == nil || ctx.final {
			return nil, errHashFinalized("hash_copy")
		}
		// A digest of the standard library marshals its state; replaying is
		// not possible, so the copy is rebuilt from the state when offered.
		type marshaler interface {
			MarshalBinary() ([]byte, error)
			UnmarshalBinary([]byte) error
		}
		copyCtx := &HashContext{algo: ctx.algo, key: ctx.key, hmac: ctx.hmac}
		factory := digests[ctx.algo]
		if ctx.hmac {
			copyCtx.h = hmac.New(factory, []byte(ctx.key))
		} else {
			copyCtx.h = factory()
		}
		if src, ok := ctx.h.(marshaler); ok {
			if state, err := src.MarshalBinary(); err == nil {
				if dst, ok := copyCtx.h.(marshaler); ok && dst.UnmarshalBinary(state) == nil {
					return copyCtx, nil
				}
			}
		}
		return nil, errHashFinalized("hash_copy")
	})
	rt.RegisterFunc("hash_file", func(algo, filename string, binary ...bool) (any, error) {
		h, err := newDigest("hash_file", algo)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			return false, nil
		}
		h.Write(data)
		return digestString(h.Sum(nil), binary), nil
	})
}

// errHashKey is the ValueError php raises for HASH_HMAC without a key.
var errHashKey = errorString("hash_init(): Argument #3 ($key) cannot be empty when HMAC is requested")

type errorString string

func (e errorString) Error() string { return string(e) }

func errHashFinalized(fn string) error {
	return errorString(fn + "(): Argument #1 ($context) must be a valid, non-finalized HashContext")
}
