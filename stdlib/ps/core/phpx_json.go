//go:build phpx

package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// init replaces upstream's json_decode with PHP's. It runs after json.go's
// init, which is why the registration here wins: the bindings apply in the
// order they were contributed and this file sorts after json.go.
func init() {
	runner.RegisterArea(`Runtime\PX\Core\JSON`, registerPhpxJSON)
}

// registerPhpxJSON installs the decoder that answers PHP's object shape.
//
// Upstream's decoder builds an *model.Array for a JSON object and refuses
// $associative = false. PHP's default answer is stdClass, and a decoded body
// read with `->name` is the shape framework code is written against. The
// extension JSON\Decoder stays as upstream built it: it has no PHP
// counterpart, and the fork changes the PHP name alone.
func registerPhpxJSON(rt *runner.Runtime) {
	rt.RegisterFunc("json_decode", phpxJSONDecode)
}

// phpxJSONDecode parses the JSON in text the way PHP does.
//
// $associative selects the shape: true answers upstream's decoder unchanged,
// and false, null or an omitted argument answers an *model.Object whose class
// is stdClass. $depth and $flags stay ignored, as upstream ignores them, and
// invalid input still raises rather than answering null.
func phpxJSONDecode(text string, opts ...any) (any, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	if len(opts) > 0 && opts[0] != nil && phpval.Truthy(opts[0]) {
		return jsonDecodeStream(dec)
	}
	return phpxJSONDecodeObject(dec)
}

// phpxJSONDecodeObject reads one value from the token stream with an object
// folded into stdClass. An array keeps upstream's []any, and each element
// takes the same shape.
//
// It reads tokens and decodes into no Go map, for the reason upstream's
// decoder does not: a map has no order and PHP keeps the document's. An
// object's properties are set in the order the document names them, so
// encoding the value again reads the way it arrived.
func phpxJSONDecodeObject(dec *json.Decoder) (any, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return jsonDecodeValue(token), nil
	}

	switch delim {
	case '{':
		out := model.NewStdClass()
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return nil, err
			}
			value, err := phpxJSONDecodeObject(dec)
			if err != nil {
				return nil, err
			}
			name, _ := key.(string)
			out.SetProp(name, value)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return out, nil
	case '[':
		out := []any{}
		for dec.More() {
			value, err := phpxJSONDecodeObject(dec)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, fmt.Errorf("json_decode(): unexpected %v", delim)
}
