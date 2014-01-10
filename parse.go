package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
)

var shebang = []byte("#!")

func Parse(usehtml bool, e, left, right string, fs map[string]interface{}, files ...string) (t template, err error) {
	new := htmlOrText[usehtml]
	if e != "" {
		t = new("", left, right, fs)
		if t, err = t.Parse(e); err != nil {
			return nil, err
		}
	}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if len(b) > 2 && bytes.Equal(b[:3], shebang) {
			if i := bytes.IndexAny(b, "\n"); i > 0 && len(b) != i {
				b = b[i+1:]
			} else {
				b = nil
			}
		}

		name := filepath.Base(file)

		var tmpl template
		if t == nil {
			t = new(name, left, right, fs)
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		if _, err = tmpl.Parse(string(b)); err != nil {
			return nil, err
		}
	}
	return
}
