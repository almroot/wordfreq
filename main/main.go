package main

import (
	"fmt"
	"os"
	"sort"
)

func main() {

	// Parse the arguments which may override the environment variables.
	var opts = NewOptions()
	var stderr = os.Stderr
	if code, terminate := opts.Parse(os.Args, stderr); terminate {
		os.Exit(code)
	}

	// Gather the files to be included...
	files, err := findFiles(opts)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		os.Exit(1)
	} else if opts.Input.ListFiles {
		sort.Strings(files)
		for _, f := range files {
			fmt.Println(f)
		}
		os.Exit(0)
	}

	// Gather the words based on frequency
	words, err := aggregate(files, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var counter int
	for _, p := range sortByWordCount(words) {
		if opts.After.ResultsFreq > 0 && p.Value < opts.After.ResultsFreq {
			break
		}
		for _, fixed := range expand(opts, p.Key) {
			if opts.After.CSV {
				fmt.Printf("%d\t%s\n", p.Value, fixed)
			} else {
				fmt.Println(fixed)
			}
		}
		counter++
		if opts.After.ResultsMax > 0 && counter > opts.After.ResultsMax {
			break
		}
	}
}
