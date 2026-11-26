package modelutil

// byRefArgs names, per builtin, the argument positions PHP passes by reference.
var byRefArgs = map[string]map[int]bool{
	"getmxrr":                     {1: true, 2: true},
	"dns_get_mx":                  {1: true, 2: true},
	"similar_text":                {2: true},
	"str_ireplace":                {3: true},
	"sscanf":                      {2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true},
	"preg_match_all":              {2: true},
	"preg_match":                  {2: true},
	"preg_replace_callback":       {4: true},
	"preg_replace":                {4: true},
	"preg_filter":                 {4: true},
	"preg_replace_callback_array": {3: true},
	"parse_str":                   {1: true},
	"curl_multi_exec":             {1: true},
	"proc_open":                   {2: true},
	"fsockopen":                   {2: true, 3: true},
	"curl_multi_info_read":        {1: true},
	"settype":                     {0: true},
	// The GCM tag an openssl call answers through a variable. The call site
	// passes the setter and the current value, so a decrypt reads the tag it
	// verifies; see phpxOpenSSLTagArgs.
	"openssl_encrypt":             {5: true},
	"openssl_decrypt":             {5: true},
	"openssl_random_pseudo_bytes": {1: true},

	// The exit status of a command. exec's $output is not here: the caller and
	// the binding share one array. The binding appends into the caller's array
	// and reaches the same observable result. An integer has no
	// such route back.
	"exec":     {2: true},
	"system":   {1: true},
	"passthru": {1: true},
}

// ByRefArg reports whether the argument at index is an output parameter of the
// named call. Either name may match; a namespaced call falls back to global.
func ByRefArg(name, fallback string, index int) bool {
	if refs, ok := byRefArgs[name]; ok {
		return refs[index]
	}
	return byRefArgs[fallback][index]
}

// arrayWriteArgs names, per builtin, the argument positions that take an array
// by reference and write it in place. The call site resolves the container
// before the call, so an array shared with a copy splits there and the write
// lands in the array the script named.
var arrayWriteArgs = map[string]map[int]bool{
	"sort":                 {0: true},
	"rsort":                {0: true},
	"usort":                {0: true},
	"uasort":               {0: true},
	"uksort":               {0: true},
	"ksort":                {0: true},
	"krsort":               {0: true},
	"asort":                {0: true},
	"arsort":               {0: true},
	"natsort":              {0: true},
	"natcasesort":          {0: true},
	"shuffle":              {0: true},
	"array_push":           {0: true},
	"array_pop":            {0: true},
	"array_shift":          {0: true},
	"array_unshift":        {0: true},
	"array_splice":         {0: true},
	"array_walk":           {0: true},
	"array_walk_recursive": {0: true},
}

// arrayWriteEveryArg names the builtins whose every array argument is written
// in place: the position is the script's, so the call site cannot name it.
var arrayWriteEveryArg = map[string]bool{
	"array_multisort": true,
}

// ArrayWriteArg reports whether the argument at index is an array the named
// builtin writes in place.
func ArrayWriteArg(name, fallback string, index int) bool {
	if arrayWriteEveryArg[name] || arrayWriteEveryArg[fallback] {
		return true
	}
	if refs, ok := arrayWriteArgs[name]; ok {
		return refs[index]
	}
	return arrayWriteArgs[fallback][index]
}
