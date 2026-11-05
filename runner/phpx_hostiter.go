package runner

// PhpxHostIterator is a Go value a script iterates with foreach: a generator.
type PhpxHostIterator interface {
	Rewind() error
	Valid() (bool, error)
	Current() (any, error)
	Key() (any, error)
	Next() error
}

// phpxForeachHost drives a host iterator through foreach.
func phpxForeachHost(it PhpxHostIterator, iter func(k, v any) bool) error {
	if err := it.Rewind(); err != nil {
		return err
	}
	for {
		ok, err := it.Valid()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		k, err := it.Key()
		if err != nil {
			return err
		}
		v, err := it.Current()
		if err != nil {
			return err
		}
		if !iter(k, v) {
			return nil
		}
		if err := it.Next(); err != nil {
			return err
		}
	}
}

// PhpxWalkHost visits every element of a host iterator.
func PhpxWalkHost(it PhpxHostIterator, visit func(k, v any) bool) error {
	return phpxForeachHost(it, visit)
}
