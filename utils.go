package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
)

func rdr(corpus string) io.Reader {
	return strings.NewReader(corpus)
}

var intSize = reflect.ValueOf(0)

func overflow(i int64) (int, error) {
	if intSize.OverflowInt(i) {
		return 0, fmt.Errorf("%d cannot be used as index on 32bit systems", i)
	}
	return int(i), nil
}

func index(index interface{}) (x int, err error) {
	switch v := reflect.ValueOf(index); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return overflow(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return overflow(int64(v.Uint()))
	default:
		return 0, fmt.Errorf("can't use type %s as index", v.Type())
	}
}

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

func hdr2map(h []string, submatch bool) (out map[string]int) {
	if len(h) == 0 {
		return map[string]int{}
	}
	x := 0
	if submatch {
		x = 1 //submatch index 0 always full line so we need to shift the indicies by one
	}
	out = make(map[string]int, len(h))
	for i, n := range h {
		if n != "" {
			out[n] = i + x
		}
	}
	return
}

func splitHeader(h string) (out []string) {
	if h == "" {
		return
	}
	for _, s := range strings.Split(h, ",") {
		out = append(out, strings.TrimSpace(s))
	}
	return
}

func oneOf(these ...bool) bool {
	for _, p := range these {
		if p {
			return true
		}
	}
	return false
}

func multiple(of ...bool) bool {
	one := false
	for _, p := range of {
		if p {
			if one {
				return true
			} else {
				one = true
			}
		}
	}
	return false
}
