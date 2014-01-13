package main

import (
	"bytes"
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
