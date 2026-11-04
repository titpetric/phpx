package runner

// PhpxHostIterator is a Go value a script iterates with foreach: a generator.
type PhpxHostIterator interface {
	Rewind() error
	Valid() bool
	Current() any
	Key() any
	Next() error
}

// phpxForeachHost drives a host iterator through foreach.
func phpxForeachHost(it PhpxHostIterator, iter func(k, v any) bool) error {
	if err := it.Rewind(); err != nil {
		return err
	}
	for it.Valid() {
		if !iter(it.Key(), it.Current()) {
			return nil
		}
		if err := it.Next(); err != nil {
			return err
		}
	}
	return nil
}

// PhpxWalkHost visits every element of a host iterator.
func PhpxWalkHost(it PhpxHostIterator, visit func(k, v any) bool) error {
	return phpxForeachHost(it, visit)
}
