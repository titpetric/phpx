package phpval

// PhpxResourceID, set by the runner, answers the id of a stream handle, which
// is what a resource is in a string or an integer context.
var PhpxResourceID func(v any) (int64, bool)

func phpxResource(v any) (int64, bool) {
	if PhpxResourceID == nil {
		return 0, false
	}
	return PhpxResourceID(v)
}
