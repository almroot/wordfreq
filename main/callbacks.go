package main

import (
	"strings"
)

func CallbackTrimSpace(_ *Options, line string) (string, error) {
	return strings.TrimSpace(line), nil
}

func CallbackTrimLeft(opts *Options, line string) (string, error) {
	if opts.Process.TrimLeft != "" {
		for _, c := range strings.Split(opts.Process.TrimLeft, ",") {
			line = strings.TrimLeft(line, c)
		}
	}
	return line, nil
}

func CallbackTrimRight(opts *Options, line string) (string, error) {
	if opts.Process.TrimRight != "" {
		for _, c := range strings.Split(opts.Process.TrimRight, ",") {
			line = strings.TrimRight(line, c)
		}
	}
	return line, nil
}

func CallbackCutRight(opts *Options, line string) (string, error) {
	if opts.Process.CutRight != "" {
		for _, c := range strings.Split(opts.Process.CutRight, ",") {
			if strings.Contains(line, c) {
				line = line[strings.Index(line, c)+1:]
			}
		}
	}
	return line, nil
}

func CallbackCutLeft(opts *Options, line string) (string, error) {
	if opts.Process.CutLeft != "" {
		for _, c := range strings.Split(opts.Process.CutLeft, ",") {
			if strings.Contains(line, c) {
				line = line[:strings.LastIndex(line, c)]
			}
		}
	}
	return line, nil
}

func CallbackCaseLower(_ *Options, line string) (string, error) {
	return strings.ToLower(line), nil
}

func CallbackCaseUpper(_ *Options, line string) (string, error) {
	return strings.ToUpper(line), nil
}
