//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxCurl)
}

// CurlHandle is the value curl_init returns. php 8 names the class CurlHandle
// and a library tests `instanceof \CurlHandle`; the Go type carries the name.
type CurlHandle struct {
	opts   map[int64]any
	errno  int64
	errmsg string
	info   map[string]any
	body   string // the RETURNTRANSFER result, for curl_multi_getcontent
	multi  *CurlMultiHandle
}

// CurlMultiHandle is the value curl_multi_init returns.
type CurlMultiHandle struct {
	handles []*CurlHandle
	pending []*CurlHandle
	done    []curlMessage
	opts    map[int64]any
}

// CurlShareHandle is the value curl_share_init returns. It shares nothing here:
// every transfer opens its own connection.
type CurlShareHandle struct {
	opts map[int64]any
}

type curlMessage struct {
	handle *CurlHandle
	result int64
}

// curlOpt resolves a CURLOPT_ name to its number.
func curlOpt(name string) int64 { return phpxCurlIntConstants[name] }

// curlCallbackOptions are the options whose value is a callable or null.
var curlCallbackOptions = map[int64]string{}

func init() {
	for _, name := range []string{"CURLOPT_HEADERFUNCTION", "CURLOPT_WRITEFUNCTION", "CURLOPT_READFUNCTION", "CURLOPT_PROGRESSFUNCTION", "CURLOPT_XFERINFOFUNCTION", "CURLOPT_DEBUGFUNCTION", "CURLOPT_PREREQFUNCTION"} {
		if n, ok := phpxCurlIntConstants[name]; ok {
			curlCallbackOptions[n] = name
		}
	}
}

// curlOptionKnown reports whether n is an option number libcurl defines.
func curlOptionKnown(n int64) bool {
	if n < 0 {
		return false
	}
	for name, v := range phpxCurlIntConstants {
		if v == n && len(name) > 8 && name[:8] == "CURLOPT_" {
			return true
		}
	}
	return false
}

// registerPhpxCurl installs the curl extension over net/http.
func registerPhpxCurl(rt *runner.Runtime) {
	for name, value := range phpxCurlIntConstants {
		rt.SetConst(name, value)
	}
	for name, value := range phpxCurlStringConstants {
		rt.SetConst(name, value)
	}
	rt.RegisterFunc("curl_init", func(rest ...any) *CurlHandle {
		h := &CurlHandle{opts: map[int64]any{}, info: map[string]any{}}
		if len(rest) > 0 && rest[0] != nil {
			h.opts[curlOpt("CURLOPT_URL")] = phpval.String(rest[0])
		}
		return h
	})
	rt.RegisterFunc("curl_setopt", func(h *CurlHandle, option int64, value any) (bool, error) {
		return h.set(rt, option, value)
	})
	rt.RegisterFunc("curl_setopt_array", func(h *CurlHandle, options any) (bool, error) {
		var failed error
		model.RangeValues(options, func(k, v any) bool {
			n, ok := k.(int64)
			if !ok {
				failed = &valueError{message: "curl_setopt_array(): Argument #2 ($options) contains an invalid cURL option"}
				return false
			}
			if _, err := h.set(rt, n, v); err != nil {
				failed = err
				return false
			}
			return true
		})
		return failed == nil, failed
	})
	rt.RegisterFunc("curl_exec", func(h *CurlHandle) (any, error) { return h.exec(rt) })
	rt.RegisterFunc("curl_getinfo", func(h *CurlHandle, rest ...any) (any, error) {
		if len(rest) > 0 && rest[0] != nil {
			return h.infoOption(phpval.Int(rest[0])), nil
		}
		out := model.NewArray()
		for _, k := range curlInfoOrder {
			if v, ok := h.info[k]; ok {
				out.Set(k, v)
			}
		}
		return out, nil
	})
	rt.RegisterFunc("curl_errno", func(h *CurlHandle) int64 { return h.errno })
	rt.RegisterFunc("curl_error", func(h *CurlHandle) string { return h.errmsg })
	rt.RegisterFunc("curl_strerror", func(code int64) any { return curlErrorText(code) })
	rt.RegisterFunc("curl_reset", func(h *CurlHandle) {
		h.opts = map[int64]any{}
		h.errno, h.errmsg, h.body = 0, "", ""
		h.info = map[string]any{}
	})
	rt.RegisterFunc("curl_close", func(h *CurlHandle) {})
	rt.RegisterFunc("curl_copy_handle", func(h *CurlHandle) *CurlHandle {
		c := &CurlHandle{opts: map[int64]any{}, info: map[string]any{}}
		for k, v := range h.opts {
			c.opts[k] = v
		}
		return c
	})
	rt.RegisterFunc("curl_pause", func(h *CurlHandle, bitmask int64) int64 { return 0 })
	rt.RegisterFunc("curl_upkeep", func(h *CurlHandle) bool { return true })
	rt.RegisterFunc("curl_escape", func(h *CurlHandle, s string) string { return phpxURLEscape(s) })
	rt.RegisterFunc("curl_unescape", func(h *CurlHandle, s string) string { return phpxURLUnescape(s) })
	rt.RegisterFunc("curl_version", func(rest ...any) any { return curlVersion() })

	rt.RegisterFunc("curl_share_init", func() *CurlShareHandle { return &CurlShareHandle{opts: map[int64]any{}} })
	rt.RegisterFunc("curl_share_setopt", func(s *CurlShareHandle, option int64, value any) bool {
		s.opts[option] = value
		return true
	})
	rt.RegisterFunc("curl_share_close", func(s *CurlShareHandle) {})
	rt.RegisterFunc("curl_share_errno", func(s *CurlShareHandle) int64 { return 0 })
	rt.RegisterFunc("curl_share_strerror", func(code int64) any { return curlShareErrorText(code) })

	rt.RegisterFunc("curl_multi_init", func() *CurlMultiHandle { return &CurlMultiHandle{opts: map[int64]any{}} })
	rt.RegisterFunc("curl_multi_add_handle", func(m *CurlMultiHandle, h *CurlHandle) int64 {
		if m == nil || h == nil {
			return 2 // CURLM_BAD_EASY_HANDLE
		}
		for _, existing := range m.handles {
			if existing == h {
				return 7 // CURLM_ADDED_ALREADY
			}
		}
		m.handles = append(m.handles, h)
		m.pending = append(m.pending, h)
		h.multi = m
		return 0
	})
	rt.RegisterFunc("curl_multi_remove_handle", func(m *CurlMultiHandle, h *CurlHandle) int64 {
		if m == nil || h == nil {
			return 2
		}
		m.handles = curlWithout(m.handles, h)
		m.pending = curlWithout(m.pending, h)
		kept := m.done[:0]
		for _, msg := range m.done {
			if msg.handle != h {
				kept = append(kept, msg)
			}
		}
		m.done = kept
		h.multi = nil
		return 0
	})
	rt.RegisterFunc("curl_multi_exec", func(m *CurlMultiHandle, still func(any)) (int64, error) {
		// A transfer runs to its end inside this call: the callbacks are
		// script code and the interpreter runs one goroutine.
		for len(m.pending) > 0 {
			h := m.pending[0]
			m.pending = m.pending[1:]
			result, err := h.exec(rt)
			if err != nil {
				return 0, err
			}
			code := int64(0)
			if result == false {
				code = h.errno
			}
			m.done = append(m.done, curlMessage{handle: h, result: code})
		}
		if still != nil {
			still(int64(0))
		}
		return 0, nil
	})
	rt.RegisterFunc("curl_multi_select", func(m *CurlMultiHandle, rest ...any) int64 {
		if len(m.done) > 0 || len(m.pending) > 0 {
			return 1
		}
		return 0
	})
	rt.RegisterFunc("curl_multi_info_read", func(m *CurlMultiHandle, queued func(any)) any {
		if len(m.done) == 0 {
			if queued != nil {
				queued(int64(0))
			}
			return false
		}
		msg := m.done[0]
		m.done = m.done[1:]
		if queued != nil {
			queued(int64(len(m.done)))
		}
		out := model.NewArray()
		out.Set("msg", int64(1))
		out.Set("result", msg.result)
		out.Set("handle", msg.handle)
		return out
	})
	rt.RegisterFunc("curl_multi_getcontent", func(h *CurlHandle) any {
		if _, ok := h.opts[curlOpt("CURLOPT_RETURNTRANSFER")]; !ok {
			return nil
		}
		return h.body
	})
	rt.RegisterFunc("curl_multi_setopt", func(m *CurlMultiHandle, option int64, value any) bool {
		m.opts[option] = value
		return true
	})
	rt.RegisterFunc("curl_multi_close", func(m *CurlMultiHandle) {
		for _, h := range m.handles {
			h.multi = nil
		}
		m.handles, m.pending, m.done = nil, nil, nil
	})
	rt.RegisterFunc("curl_multi_errno", func(m *CurlMultiHandle) int64 { return 0 })
	rt.RegisterFunc("curl_multi_strerror", func(code int64) any { return curlMultiErrorText(code) })
}

func curlWithout(list []*CurlHandle, h *CurlHandle) []*CurlHandle {
	out := list[:0]
	for _, x := range list {
		if x != h {
			out = append(out, x)
		}
	}
	return out
}

// set stores one option, refusing a number libcurl does not define and a
// callback option whose value is not callable.
func (h *CurlHandle) set(rt *runner.Runtime, option int64, value any) (bool, error) {
	if !curlOptionKnown(option) {
		return false, &valueError{message: "curl_setopt(): Argument #2 ($option) is not a valid cURL option"}
	}
	if name, isCallback := curlCallbackOptions[option]; isCallback && value != nil {
		if _, ok := rt.Callable(value); !ok {
			return false, &curlTypeError{message: "curl_setopt(): The " + name + " option must have a valid callback"}
		}
	}
	if value == nil {
		delete(h.opts, option)
		return true, nil
	}
	if name, isList := curlListOptions[option]; isList {
		if _, ok := value.(*model.Array); !ok {
			return false, &curlTypeError{message: "curl_setopt(): The " + name + " option must have an array value"}
		}
	} else if obj, isObject := value.(*model.Object); isObject && curlStringOption(option) {
		text, err := phpval.StringContext(obj)
		if err != nil {
			return false, err
		}
		value = text
	}
	h.opts[option] = value
	return true, nil
}

// curlListOptions are the options whose value must be an array.
var curlListOptions = map[int64]string{}

// curlNonStringOptions are the string-range options that take a stream, an
// array or any value, which an object may stand in for unconverted.
var curlNonStringOptions = map[int64]bool{}

func init() {
	for _, name := range []string{"CURLOPT_HTTPHEADER", "CURLOPT_PROXYHEADER", "CURLOPT_RESOLVE", "CURLOPT_CONNECT_TO", "CURLOPT_QUOTE", "CURLOPT_POSTQUOTE", "CURLOPT_PREQUOTE", "CURLOPT_HTTP200ALIASES", "CURLOPT_MAIL_RCPT", "CURLOPT_TELNETOPTIONS"} {
		if n, ok := phpxCurlIntConstants[name]; ok {
			curlListOptions[n] = name
		}
	}
	for _, name := range []string{"CURLOPT_POSTFIELDS", "CURLOPT_FILE", "CURLOPT_INFILE", "CURLOPT_STDERR", "CURLOPT_WRITEHEADER", "CURLOPT_PRIVATE", "CURLOPT_SHARE"} {
		if n, ok := phpxCurlIntConstants[name]; ok {
			curlNonStringOptions[n] = true
		}
	}
}

// curlStringOption reports whether option takes a string, which php converts
// from a Stringable object and refuses from any other.
func curlStringOption(option int64) bool {
	return option >= 10000 && option < 20000 && curlListOptions[option] == "" && !curlNonStringOptions[option]
}

// curlTypeError is the TypeError php raises for a bad option value.
type curlTypeError struct{ message string }

func (e *curlTypeError) Error() string          { return e.message }
func (e *curlTypeError) ThrowableClass() string { return "TypeError" }
