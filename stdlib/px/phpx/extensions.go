//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Extensions`, registerExtensions)
}

// extensionMarker names the symbol that proves an extension's surface is
// present. A function is the usual proof; a class covers an extension whose
// surface is a class, such as dom's DOMDocument.
type extensionMarker struct {
	symbol string
	class  bool
}

// extensionMarkers names, per extension, the symbol whose presence means the
// runtime carries that extension's surface. An empty marker is always loaded.
var extensionMarkers = map[string]extensionMarker{
	"core": {}, "standard": {}, "spl": {}, "date": {}, "reflection": {},
	"mbstring":  {symbol: "mb_strlen"},
	"ctype":     {symbol: "ctype_alpha"},
	"json":      {symbol: "json_encode"},
	"pcre":      {symbol: "preg_match"},
	"hash":      {symbol: "hash"},
	"zlib":      {symbol: "gzcompress"},
	"curl":      {symbol: "curl_init"},
	"filter":    {symbol: "filter_var"},
	"iconv":     {symbol: "iconv"},
	"tokenizer": {symbol: "token_get_all"},
	"openssl":   {symbol: "openssl_encrypt"},
	"pdo":       {symbol: "PDO", class: true},
	"dom":       {symbol: "DOMDocument", class: true},
	"xmlwriter": {symbol: "XMLWriter", class: true},
}

func registerExtensions(rt *runner.Runtime) {
	hasClass := func(name string) bool {
		ok, _ := rt.ClassExists(name, false)
		return ok
	}
	rt.RegisterFunc("extension_loaded", func(name string) bool {
		return extensionLoaded(rt.FunctionExists, hasClass, name)
	})
}

// extensionLoaded answers extension_loaded for name with exists and hasClass
// as the probes.
func extensionLoaded(exists, hasClass func(string) bool, name string) bool {
	marker, ok := extensionMarkers[strings.ToLower(name)]
	if !ok {
		return false
	}
	if marker.symbol == "" {
		return true
	}
	if marker.class {
		return hasClass(marker.symbol)
	}
	return exists(marker.symbol)
}
