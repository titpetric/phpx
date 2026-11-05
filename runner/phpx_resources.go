//go:build phpx

package runner

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
)

// A php resource is a Go stream here: anything that reads or writes. Its id is
// a small number handed out the first time the handle is named, and a closed
// handle is remembered as closed, because the Go value is still a value.
var phpxResources struct {
	ids    sync.Map // handle -> int64
	closed sync.Map // handle -> struct{}
	next   atomic.Int64
}

func init() { phpxResources.next.Store(4) }

// PhpxIsResource reports whether v is a stream handle, open or closed.
func PhpxIsResource(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case string, bool, int, int64, float64:
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer && rv.Kind() != reflect.Interface && rv.Kind() != reflect.Struct {
		return false
	}
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return false
	}
	_, reads := v.(io.Reader)
	_, writes := v.(io.Writer)
	return reads || writes
}

// PhpxResourceID answers the resource id of a handle.
func PhpxResourceID(v any) int64 {
	if id, ok := phpxResources.ids.Load(v); ok {
		return id.(int64)
	}
	id := phpxResources.next.Add(1)
	actual, _ := phpxResources.ids.LoadOrStore(v, id)
	return actual.(int64)
}

// PhpxCloseResource remembers that fclose() closed the handle.
func PhpxCloseResource(v any) { phpxResources.closed.Store(v, struct{}{}) }

// PhpxResourceClosed reports whether the handle was closed.
func PhpxResourceClosed(v any) bool {
	_, closed := phpxResources.closed.Load(v)
	return closed
}

// phpxResourceString is what a resource becomes in a string context.
func phpxResourceString(v any) string { return fmt.Sprintf("Resource id #%d", PhpxResourceID(v)) }
