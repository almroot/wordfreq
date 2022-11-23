package main

// Ensure with run a single callback over and over until the result no longer changes
type Ensure struct {
	callback func(opts *Options, line string) (string, error)
}

// NewEnsure constructs an Ensure instance having a specific callback
func NewEnsure(callback func(opts *Options, line string) (string, error)) *Ensure {
	return &Ensure{callback: callback}
}

// Callback is the implementation which will run the inner callback continuously until the results
// no longer mutates. This will allow us to, for example, trim a string multiple times.
func (e *Ensure) Callback(opts *Options, line string) (string, error) {
	previous := line
	for {
		out, err := e.callback(opts, previous)
		if err != nil {
			return previous, err
		} else if out == previous {
			return out, nil
		}
		previous = out
	}
}
