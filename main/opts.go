package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobwas/glob"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Core    *OptionsCore
	Input   *OptionsInput
	Before  *OptionsBefore
	Process *OptionsDuring
	After   *OptionsAfter
}

type OptionsCore struct {
}

type OptionsInput struct {
	Wordlist     string `short:"w" long:"wordlist" description:"The location where we are to consume wordlists from"`
	IgnoreCase   bool   `short:"i" long:"ignore" description:"Ignore filename casing"`
	ListFiles    bool   `short:"l" long:"list" description:"List the files that are to be processed"`
	Include      string `long:"include" description:"Glob pattern of relative file names to include"`
	Exclude      string `long:"exclude" description:"Glob pattern of relative file names to exclude"`
	Bias         string `long:"bias" description:"A comma-separated string of factor:glob, the factor modifies the default score of 1"`
	GlobsInclude []glob.Glob
	GlobsExclude []glob.Glob
	GlobsBias    map[glob.Glob]float64
}

type OptionsBefore struct {
	GlobsInclude []glob.Glob
	GlobsExclude []glob.Glob
	RegexInclude []*regexp.Regexp
	RegexExclude []*regexp.Regexp
	IncludeGlob  string `long:"pre-include-glob" description:"Glob pattern of included matches"`
	ExcludeGlob  string `long:"pre-exclude-glob" description:"Glob pattern of excluded matches"`
	IncludeRegex string `long:"pre-include-regex" description:"Regex pattern of included matches"`
	ExcludeRegex string `long:"pre-exclude-regex" description:"Regex pattern of included matches"`
}

type OptionsDuring struct {
	Callbacks []func(opts *Options, line string) (string, error)
	Order     string `long:"order" description:"The order in which directives are carried out, see supported opcodes below"`
	TrimSpace bool   `short:"e" description:"Trims whitespaces before and after the string"`
	TrimLeft  string `short:"a" description:"Characters to remove from the prefix"`
	TrimRight string `short:"f" description:"Characters to remove from the suffix"`
	CutLeft   string `short:"d" description:"Cuts the string based on the delimiter(s) and keep the prefix"`
	CutRight  string `short:"s" description:"Cuts the string based on the delimiter(s) and keep the suffix"`
	CaseLower bool   `short:"l" description:"Normalizes the string to lowercase"`
	CaseUpper bool   `short:"u" description:"Normalizes the string to uppercase"`
}

type OptionsAfter struct {
	GlobsInclude []glob.Glob
	GlobsExclude []glob.Glob
	RegexInclude []*regexp.Regexp
	RegexExclude []*regexp.Regexp
	IncludeGlob  string  `long:"post-include-glob" description:"Glob pattern of included matches"`
	ExcludeGlob  string  `long:"post-exclude-glob" description:"Glob pattern of excluded matches"`
	IncludeRegex string  `long:"post-include-regex" description:"Regex pattern of included matches"`
	ExcludeRegex string  `long:"post-exclude-regex" description:"Regex pattern of included matches"`
	ValuePrefix  string  `long:"value-prefix" description:"A set of strings to prepend to the final string"`
	ValueSuffix  string  `long:"value-suffix" description:"A set of strings to append to the final string"`
	ResultsMax   int     `long:"results-max" description:"The amount of results to return"`
	ResultsFreq  float64 `long:"results-freq" description:"The cut off rate on frequency for which we will abort"`
	CSV          bool    `long:"csv" description:"Produces a CSV separated by tab, having the frequency and word"`
}

func NewOptions() *Options {
	return &Options{
		Core: &OptionsCore{},
		Input: &OptionsInput{
			Wordlist: DefaultDirectory(),
			Exclude:  ".git",
			Bias:     "1.0:*",
		},
		Before: &OptionsBefore{},
		Process: &OptionsDuring{
			Order: "EfaDSfaE",
		},
		After: &OptionsAfter{},
	}
}

func DefaultDirectory() string {
	const loc = "wordlists"
	var path = os.TempDir()
	if override, err := os.UserHomeDir(); err == nil {
		path = override
	}
	return filepath.Join(path, loc)
}

func (o *Options) Parse(args []string, stderr io.Writer) (int, bool) {
	const errorFormat = "error: %v\n"
	var parser = flags.NewParser(nil, flags.Default)
	if _, err := parser.AddGroup("Application", "", o.Core); err != nil {
		panic(err)
	} else if _, err = parser.AddGroup("Input", "", o.Input); err != nil {
		panic(err)
	} else if _, err = parser.AddGroup("Pre-Processing", "", o.Before); err != nil {
		panic(err)
	} else if _, err = parser.AddGroup("Processing", "", o.Process); err != nil {
		panic(err)
	} else if _, err = parser.AddGroup("Post-Processing", "", o.After); err != nil {
		panic(err)
	}
	_, err := parser.ParseArgs(args)
	if flags.WroteHelp(err) {
		return 0, true
	} else if err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err = o.Core.Parse(o); err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err = o.Input.Parse(o); err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err = o.Before.Parse(o); err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err = o.Process.Parse(o); err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err = o.After.Parse(o); err != nil {
		_, _ = fmt.Fprintf(stderr, errorFormat, err)
		return 0, true
	} else if err == nil {
		return 0, false
	}
	return 1, true
}

func (x *OptionsCore) Parse(o *Options) error {
	return nil
}

func (x *OptionsInput) Parse(o *Options) error {
	if info, err := os.Stat(x.Wordlist); err != nil {
		return err
	} else if !info.IsDir() {
		const msg = "not a directory"
		return errors.New(msg)
	}
	include, err := parseGlobs(x.Include)
	if err != nil {
		return err
	}
	exclude, err := parseGlobs(x.Exclude)
	if err != nil {
		return err
	}
	x.GlobsInclude = include
	x.GlobsExclude = exclude
	x.GlobsBias = make(map[glob.Glob]float64)
	for _, c := range strings.Split(x.Bias, ",") {
		k, v, found := strings.Cut(c, ":")
		if !found {
			return errors.New("lacking colon")
		}
		var factor float64
		var instance glob.Glob
		if factor, err = strconv.ParseFloat(k, 64); err != nil {
			return err
		} else if instance, err = glob.Compile(v); err != nil {
			return err
		} else {
			x.GlobsBias[instance] = factor
		}
	}
	return nil
}

func (x *OptionsBefore) Parse(o *Options) error {
	includeGlob, err := parseGlobs(x.IncludeGlob)
	if err != nil {
		return err
	}
	excludeGlob, err := parseGlobs(x.ExcludeGlob)
	if err != nil {
		return err
	}
	x.GlobsInclude = includeGlob
	x.GlobsExclude = excludeGlob
	includeRegex, err := parseRegex(x.IncludeRegex)
	if err != nil {
		return err
	}
	excludeRegex, err := parseRegex(x.ExcludeRegex)
	if err != nil {
		return err
	}
	x.RegexInclude = includeRegex
	x.RegexExclude = excludeRegex
	return nil
}

func (x *OptionsDuring) Parse(o *Options) error {
	if x.CaseUpper && !strings.ContainsAny(x.Order, "uU") {
		x.Order += "l"
	}
	if x.CaseUpper && !strings.ContainsAny(x.Order, "uU") {
		x.Order += "u"
	}
	if x.TrimSpace && !strings.ContainsAny(x.Order, "eE") {
		x.Order += "e"
	}
	for idx, d := range x.Order {
		switch d {
		case 'e':
			x.Callbacks = append(x.Callbacks, CallbackTrimSpace)
		case 'E':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackTrimSpace).Callback)
		case 'a':
			x.Callbacks = append(x.Callbacks, CallbackTrimLeft)
		case 'A':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackTrimLeft).Callback)
		case 'f':
			x.Callbacks = append(x.Callbacks, CallbackTrimRight)
		case 'F':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackTrimRight).Callback)
		case 's':
			x.Callbacks = append(x.Callbacks, CallbackCutRight)
		case 'S':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackCutRight).Callback)
		case 'd':
			x.Callbacks = append(x.Callbacks, CallbackCutLeft)
		case 'D':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackCutLeft).Callback)
		case 'l':
			x.Callbacks = append(x.Callbacks, CallbackCaseLower)
		case 'L':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackCaseLower).Callback)
		case 'u':
			x.Callbacks = append(x.Callbacks, CallbackCaseUpper)
		case 'U':
			x.Callbacks = append(x.Callbacks, NewEnsure(CallbackCaseUpper).Callback)
		default:
			return fmt.Errorf("unrecognized directive 0x%x", x.Order[idx])
		}
	}
	return nil
}

func (x *OptionsAfter) Parse(o *Options) error {
	includeGlob, err := parseGlobs(x.IncludeGlob)
	if err != nil {
		return err
	}
	excludeGlob, err := parseGlobs(x.ExcludeGlob)
	if err != nil {
		return err
	}
	x.GlobsInclude = includeGlob
	x.GlobsExclude = excludeGlob
	includeRegex, err := parseRegex(x.IncludeRegex)
	if err != nil {
		return err
	}
	excludeRegex, err := parseRegex(x.ExcludeRegex)
	if err != nil {
		return err
	}
	x.RegexInclude = includeRegex
	x.RegexExclude = excludeRegex
	return nil
}
