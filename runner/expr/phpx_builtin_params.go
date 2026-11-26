package expr

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// exceptionParams are the constructor parameters of every built-in throwable.
var exceptionParams = []string{"message", "code", "previous"}

// builtinParams names the parameters of the built-in functions and classes a
// named argument reaches. A Go binding records no parameter names, so a call
// with named arguments is put into positional order from this table when it is
// compiled. Keys are lower case; a class is its name.
var builtinParams = map[string][]string{
	"exception": exceptionParams, "error": exceptionParams, "runtimeexception": exceptionParams,
	"logicexception": exceptionParams, "invalidargumentexception": exceptionParams,
	"unexpectedvalueexception": exceptionParams, "outofboundsexception": exceptionParams,
	"outofrangeexception": exceptionParams, "badmethodcallexception": exceptionParams,
	"badfunctioncallexception": exceptionParams, "domainexception": exceptionParams,
	"lengthexception": exceptionParams, "rangeexception": exceptionParams,
	"overflowexception": exceptionParams, "underflowexception": exceptionParams,
	"typeerror": exceptionParams, "valueerror": exceptionParams,
	"arithmeticerror": exceptionParams, "divisionbyzeroerror": exceptionParams,
	"jsonexception": exceptionParams,

	"str_pad":              {"string", "length", "pad_string", "pad_type"},
	"implode":              {"separator", "array"},
	"explode":              {"separator", "string", "limit"},
	"json_encode":          {"value", "flags", "depth"},
	"json_decode":          {"json", "associative", "depth", "flags"},
	"in_array":             {"needle", "haystack", "strict"},
	"array_search":         {"needle", "haystack", "strict"},
	"array_keys":           {"array", "filter_value", "strict"},
	"array_filter":         {"array", "callback", "mode"},
	"array_map":            {"callback", "array"},
	"array_slice":          {"array", "offset", "length", "preserve_keys"},
	"array_column":         {"array", "column_key", "index_key"},
	"array_unique":         {"array", "flags"},
	"array_key_exists":     {"key", "array"},
	"array_reverse":        {"array", "preserve_keys"},
	"array_fill":           {"start_index", "count", "value"},
	"count":                {"value", "mode"},
	"htmlspecialchars":     {"string", "flags", "encoding", "double_encode"},
	"debug_backtrace":      {"options", "limit"},
	"number_format":        {"num", "decimals", "decimal_separator", "thousands_separator"},
	"round":                {"num", "precision", "mode"},
	"substr":               {"string", "offset", "length"},
	"mb_substr":            {"string", "start", "length", "encoding"},
	"strpos":               {"haystack", "needle", "offset"},
	"str_contains":         {"haystack", "needle"},
	"str_starts_with":      {"haystack", "needle"},
	"str_ends_with":        {"haystack", "needle"},
	"str_replace":          {"search", "replace", "subject", "count"},
	"str_repeat":           {"string", "times"},
	"trim":                 {"string", "characters"},
	"rtrim":                {"string", "characters"},
	"ltrim":                {"string", "characters"},
	"ucfirst":              {"string"},
	"var_export":           {"value", "return"},
	"print_r":              {"value", "return"},
	"preg_match":           {"pattern", "subject", "matches", "flags", "offset"},
	"preg_replace":         {"pattern", "replacement", "subject", "limit", "count"},
	"is_a":                 {"object_or_class", "class", "allow_string"},
	"is_subclass_of":       {"object_or_class", "class", "allow_string"},
	"iterator_to_array":    {"iterator", "preserve_keys"},
	"call_user_func":       {"callback"},
	"sprintf":              {"format"},
	"intdiv":               {"num1", "num2"},
	"max":                  {"value"},
	"min":                  {"value"},
	"strtr":                {"string", "from", "to"},
	"str_getcsv":           {"string", "separator", "enclosure", "escape"},
	"wordwrap":             {"string", "width", "break", "cut_long_words"},
	"nl2br":                {"string", "use_xhtml"},
	"array_pad":            {"array", "length", "value"},
	"array_chunk":          {"array", "length", "preserve_keys"},
	"array_combine":        {"keys", "values"},
	"array_fill_keys":      {"keys", "value"},
	"range":                {"start", "end", "step"},
	"usort":                {"array", "callback"},
	"uasort":               {"array", "callback"},
	"uksort":               {"array", "callback"},
	"array_walk":           {"array", "callback", "arg"},
	"array_reduce":         {"array", "callback", "initial"},
	"call_user_func_array": {"callback", "args"},
}

// builtinDefaults answers the default of a named parameter a call left out, so
// `str_getcsv($line, escape: '\\')` keeps php's separator and enclosure in
// their places. A parameter with no entry and no value makes phpxOrderNamed
// decline the call: php would answer a TypeError for it, and the binding's own
// arity check says the same.
var builtinDefaults = map[string]map[string]any{
	"exception":                {"code": int64(0), "previous": nil},
	"error":                    {"code": int64(0), "previous": nil},
	"typeerror":                {"code": int64(0), "previous": nil},
	"runtimeexception":         {"code": int64(0), "previous": nil},
	"invalidargumentexception": {"code": int64(0), "previous": nil},
	"str_getcsv":               {"separator": ",", "enclosure": `"`, "escape": `\`},
	"str_pad":                  {"pad_string": " ", "pad_type": int64(1)},
	"implode":                  {"separator": ""},
	"explode":                  {"limit": int64(9223372036854775807)},
	"json_encode":              {"flags": int64(0), "depth": int64(512)},
	"json_decode":              {"associative": nil, "depth": int64(512), "flags": int64(0)},
	"in_array":                 {"strict": false},
	"array_search":             {"strict": false},
	"array_keys":               {"filter_value": nil, "strict": false},
	"array_filter":             {"callback": nil, "mode": int64(0)},
	"array_slice":              {"length": nil, "preserve_keys": false},
	"array_column":             {"index_key": nil},
	"array_unique":             {"flags": int64(2)},
	"array_reverse":            {"preserve_keys": false},
	"array_chunk":              {"preserve_keys": false},
	"count":                    {"mode": int64(0)},
	"debug_backtrace":          {"options": int64(1), "limit": int64(0)},
	"number_format":            {"decimals": int64(0), "decimal_separator": ".", "thousands_separator": ","},
	"round":                    {"precision": int64(0), "mode": int64(1)},
	"substr":                   {"length": nil},
	"mb_substr":                {"length": nil, "encoding": nil},
	"strpos":                   {"offset": int64(0)},
	"trim":                     {"characters": " \n\r\t\v\x00"},
	"rtrim":                    {"characters": " \n\r\t\v\x00"},
	"ltrim":                    {"characters": " \n\r\t\v\x00"},
	"var_export":               {"return": false},
	"print_r":                  {"return": false},
	"preg_match":               {"flags": int64(0), "offset": int64(0)},
	"preg_replace":             {"limit": int64(-1)},
	"is_a":                     {"allow_string": true},
	"is_subclass_of":           {"allow_string": true},
	"iterator_to_array":        {"preserve_keys": true},
	"wordwrap":                 {"break": "\n", "cut_long_words": false},
	"nl2br":                    {"use_xhtml": true},
	"range":                    {"step": int64(1)},
	"array_walk":               {"arg": nil},
	"array_reduce":             {"initial": nil},
	"htmlspecialchars":         {"flags": int64(11), "encoding": nil, "double_encode": true},
}

// phpxOrderNamed puts the arguments of a call to a built-in into positional
// order, or returns them as they came when the callee is not in the table, a
// name is unknown, or the table leaves a position nobody gave a value for.
func phpxOrderNamed(callee string, args []model.Expr) []model.Expr {
	named := false
	for _, a := range args {
		if _, ok := namedName(a); ok {
			named = true
		}
	}
	if !named {
		return args
	}
	// An unqualified call inside a namespace spells its own namespace on the
	// name and resolves to the global built-in when that namespace declares
	// no function of that name. The table is read by the base name, which is
	// the name the runtime falls back to.
	name := strings.ToLower(baseCallName(strings.TrimPrefix(callee, "\\")))
	params, ok := builtinParams[name]
	if !ok {
		return args
	}
	out := make([]model.Expr, len(params))
	top := 0
	for i, a := range args {
		name, isNamed := namedName(a)
		at := i
		if isNamed {
			at = -1
			for j, p := range params {
				if p == name {
					at = j
				}
			}
			a = a.(*model.Unary).X
		}
		if at < 0 || at >= len(out) || out[at] != nil {
			return args
		}
		out[at] = a
		top = max(top, at+1)
	}
	out = out[:top]
	defaults := builtinDefaults[name]
	for i, a := range out {
		if a != nil {
			continue
		}
		if i >= len(params) {
			continue
		}
		def, ok := defaults[params[i]]
		if !ok {
			// A parameter the table gives no default is null here, which the
			// binding refuses with its own arity or type message.
			def = nil
		}
		out[i] = &model.Lit{Value: def, Raw: "null"}
	}
	return out
}
