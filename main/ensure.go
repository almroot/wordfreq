package main

type Ensure struct {
	callback func(opts *Options, line string) (string, error)
}

func NewEnsure(callback func(opts *Options, line string) (string, error)) *Ensure {
	return &Ensure{callback: callback}
}

func (e *Ensure) Callback(opts *Options, line string) (string, error) {
	var previous = line
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
