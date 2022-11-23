package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

func parseGlobs(directive string) ([]glob.Glob, error) {
	const special = "*?{["
	if directive == "" {
		return nil, nil
	}
	var out []glob.Glob
	for _, chunk := range strings.Split(directive, ",") {
		var fix = true
		for _, s := range special {
			if strings.ContainsRune(chunk, s) {
				fix = false
				break
			}
		}
		if fix {
			chunk = fmt.Sprintf("*%s*", chunk)
		}
		instance, err := glob.Compile(chunk)
		if err != nil {
			return out, err
		}
		out = append(out, instance)
	}
	return out, nil
}

func parseRegex(directive string) ([]*regexp.Regexp, error) {
	if directive == "" {
		return nil, nil
	}
	var out []*regexp.Regexp
	for _, chunk := range strings.Split(directive, ",") {
		instance, err := regexp.Compile(chunk)
		if err != nil {
			return out, err
		}
		out = append(out, instance)
	}
	return out, nil
}
