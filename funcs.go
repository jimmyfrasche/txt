package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

var funcs = map[string]interface{}{
	"slice": func(what interface{}, slice ...interface{}) (interface{}, error) {
		v := reflect.ValueOf(what)
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.String:
		default:
			return nil, fmt.Errorf("can't index item of type %s", v.Type())
		}
		if ln := len(slice); ln == 0 || ln > 2 {
			return nil, fmt.Errorf("slice takes 2 or 3 parameters, given %d", ln)
		}

		ln := v.Len()
		start, err := index(slice[0])
		if err != nil {
			return nil, err
		}

		usestop := false
		stop := ln
		if len(slice) == 2 {
			usestop = true
			stop, err = index(slice[1])
			if err != nil {
				return nil, err
			}
		}

		if start == stop {
			return v.Slice(0, 0), nil
		}

		if usestop && start < 0 && stop < 0 {
			if start > stop {
				return nil, fmt.Errorf("invalid indicies (%d, %d)", start, stop)
			}
			//flip so correct when we un-negative them
			start, stop = stop, start
		}

		if start < 0 {
			start = ln + start
		}
		if usestop && stop < 0 {
			stop = ln + stop
		}

		if start > stop {
			if usestop {
				return nil, fmt.Errorf("invalid indicies (%d, %d)", start, stop)
			} else {
				return v.Slice(0, 0), nil
			}
		}

		//clamp
		if usestop && stop > ln {
			stop = ln
		}

		return v.Slice(start, stop), nil
	},
	"nl": func(s string) string {
		if len(s) == 0 || s[len(s)-1] == '\n' {
			return s
		}
		return s + "\n"
	},
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
	"readLine": func(RS, LP, header, file string) (interface{}, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if RS == "" {
			RS = *RecordSeparator
		}
		if LP == "" {
			LP = *LinePattern
		}
		hdr := splitHeader(header)
		return SubmatchSplit(hdr, RS, LP, f)
	},
	"read": func(RS, FS, header, file string) (interface{}, error) {
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
		hdr := splitHeader(header)
		return Split(hdr, RS, FS, f)
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

	"quoteGo":      strconv.Quote,
	"quoteGoASCII": strconv.QuoteToASCII,

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
