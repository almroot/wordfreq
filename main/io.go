package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
)

func findFiles(opts *Options) ([]string, error) {
	var files []string
	err := filepath.Walk(opts.Input.Wordlist, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// Find the base path of the filename...
		base, err := filepath.Rel(opts.Input.Wordlist, path)
		if err != nil {
			return nil
		} else if !isWantedGlob(base, opts.Input.GlobsInclude, opts.Input.GlobsExclude) {
			return nil
		} else if opts.Input.IgnoreCase && !isWantedGlob(strings.ToLower(base), opts.Input.GlobsInclude, opts.Input.GlobsExclude) {
			return nil
		}

		// The file is wanted! Add to the collection...
		files = append(files, path)
		return nil
	})
	return files, err
}

func isWantedGlob(data string, include, exclude []glob.Glob) bool {
	var doExclude = false
	var doInclude = len(include) == 0
	for _, i := range include {
		if i.Match(data) {
			doInclude = true
			break
		}
	}
	for _, e := range exclude {
		if e.Match(data) {
			doExclude = true
			break
		}
	}
	if doExclude {
		return false
	}
	return doInclude
}

func isWantedRegex(data string, include, exclude []*regexp.Regexp) bool {
	var doExclude = false
	var doInclude = len(include) == 0
	for _, i := range include {
		if i.MatchString(data) {
			doInclude = true
			break
		}
	}
	for _, e := range exclude {
		if e.MatchString(data) {
			doExclude = true
			break
		}
	}
	if doExclude {
		return false
	}
	return doInclude
}

func aggregate(files []string, opts *Options, stderr *os.File) (map[string]int, error) {
	var output = make(map[string]int)
	for _, f := range files {

		// Open the file having many words and start iterating line-by-line.
		file, err := os.Open(f)
		if err != nil {
			return output, err
		}
		var line string
		var scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			line = scanner.Text()

			// Start with the pre-processing directives...
			if !isWantedGlob(line, opts.Before.GlobsInclude, opts.Before.GlobsExclude) {
				continue
			} else if !isWantedRegex(line, opts.Before.RegexInclude, opts.Before.RegexExclude) {
				continue
			}

			// Then we iterate over the given callbacks following the --order flag.
			for _, cb := range opts.Process.Callbacks {
				if line, err = cb(opts, line); err != nil {
					const msg = "error processing: \"%s\" because: %v"
					_, _ = fmt.Fprintf(stderr, msg, strconv.Quote(line), err)
					break
				}
			}
			if err != nil {
				continue
			}

			// Finally we do the (most significant) post-processing steps
			for _, fixed := range expand(opts, line) {
				if !isWantedGlob(fixed, opts.After.GlobsInclude, opts.After.GlobsExclude) {
					continue
				} else if !isWantedRegex(fixed, opts.After.RegexInclude, opts.After.RegexExclude) {
					continue
				}
				output[fixed]++
			}
		}

		// Close the file handle once we are done reading.
		if err = file.Close(); err != nil {
			const msg = "error closing file %s: %v"
			_, _ = fmt.Fprintf(stderr, msg, f, err)
		}
	}
	return output, nil
}

func expand(options *Options, line string) []string {
	prefix := strings.Split(options.After.ValuePrefix, ",")
	suffix := strings.Split(options.After.ValueSuffix, ",")
	if len(prefix) == 0 {
		prefix = []string{}
	}
	if len(suffix) == 0 {
		suffix = []string{}
	}
	var out []string
	for _, pre := range prefix {
		for _, post := range suffix {
			out = append(out, fmt.Sprintf("%s%s%s", pre, line, post))
		}
	}
	return out
}
