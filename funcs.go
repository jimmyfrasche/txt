package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func swapArgs(f func(string, string) string) func(string, string) string {
	return func(a, b string) string {
		return f(b, a)
	}
}

var rc = map[string]*regexp.Regexp{}

func cmpl(p string) (*regexp.Regexp, error) {
	if r, ok := rc[p]; ok {
		return r, nil
	}
	r, err := regexp.Compile(p)
	if err != nil {
		return nil, err
	}
	rc[p] = r
	return r, nil
}

func run(c *exec.Cmd) string {
	var out bytes.Buffer
	c.Stdout = &out
	_ = c.Run()
	return out.String()
}

var funcs = map[string]interface{}{
	"readCSV": func(header, file string) (interface{}, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return CSV(strings.Split(header, ","), f)
	},
	"readJSON": func(file string) (interface{}, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return JSON(f)
	},
	"read": func(RS, FS, file string) (interface{}, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if RS == "" {
			RS = *RecordSeparator
		}
		if FS == "" {
			FS = *FieldSeparator
		}
		return Split(RS, FS, f)
	},
	"quoteCSV": func(s string) string {
		hasQuote := strings.Index(s, `"`) > 0
		hasComma := strings.Index(s, ",") > 0
		if hasComma && !hasQuote {
			return `"` + s + `"`
		}
		if hasQuote {
			return `"` + strings.Replace(s, `"`, `""`, -1) + `"`
		}
		return s
	},

	"toJSON": json.Marshal,

	"readFile": ioutil.ReadFile,

	"equalFold": strings.EqualFold,
	"fields":    strings.Fields,
	"Join": func(sep string, a []string) string {
		return strings.Join(a, sep)
	},
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"title":      strings.ToTitle,
	"trim":       swapArgs(strings.Trim),
	"trimLeft":   swapArgs(strings.TrimLeft),
	"trimRight":  swapArgs(strings.TrimRight),
	"trimPrefix": swapArgs(strings.TrimPrefix),
	"trimSuffix": swapArgs(strings.TrimSuffix),
	"trimSpace":  strings.TrimSpace,

	"match": func(pattern, src string) (bool, error) {
		r, err := cmpl(pattern)
		if err != nil {
			return false, err
		}
		return r.MatchString(src), nil
	},
	"find": func(pattern, src string) ([]string, error) {
		r, err := cmpl(pattern)
		if err != nil {
			return nil, err
		}
		return r.FindAllString(src, -1), nil
	},
	"replace": func(pattern, template, src string) (string, error) {
		r, err := cmpl(pattern)
		if err != nil {
			return "", err
		}
		return r.ReplaceAllString(src, template), nil
	},
	"split": func(pattern, src string) ([]string, error) {
		r, err := cmpl(pattern)
		if err != nil {
			return nil, err
		}
		return r.Split(src, -1), nil
	},

	"env": os.Getenv,

	"exec": func(name string, args ...string) string {
		return run(exec.Command(name, args...))
	},
	"pipe": func(name string, args ...string) (string, error) {
		if len(args) == 0 {
			return "", errors.New("pipe requires an input as the last argument")
		}
		last := len(args) - 1
		input := args[last]
		args = args[:last]
		cmd := exec.Command(name, args...)
		cmd.Stdin = strings.NewReader(input)
		return run(cmd), nil
	},
}
