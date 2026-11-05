//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Extensions`, registerExtensions)
}

// extensionMarkers names, per extension, a function whose presence means the
// runtime carries that extension's surface. An empty marker is always loaded.
var extensionMarkers = map[string]string{
	"core": "", "standard": "", "spl": "", "date": "", "reflection": "",
	"mbstring":  "mb_strlen",
	"ctype":     "ctype_alpha",
	"json":      "json_encode",
	"pcre":      "preg_match",
	"hash":      "hash",
	"zlib":      "gzcompress",
	"curl":      "curl_init",
	"filter":    "filter_var",
	"iconv":     "iconv",
	"tokenizer": "token_get_all",
	"openssl":   "openssl_encrypt",
}

func registerExtensions(rt *runner.Runtime) {
	rt.RegisterFunc("extension_loaded", func(name string) bool {
		return extensionLoaded(rt.FunctionExists, name)
	})
}

// extensionLoaded answers extension_loaded for name with exists as the probe.
func extensionLoaded(exists func(string) bool, name string) bool {
	marker, ok := extensionMarkers[strings.ToLower(name)]
	if !ok {
		return false
	}
	return marker == "" || exists(marker)
}
