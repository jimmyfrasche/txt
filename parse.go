package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

var shebang = []byte("#!")

func Parse(e, left, right string, fs template.FuncMap, files ...string) (t *template.Template, err error) {
	if e != "" {
		t = template.New("").Funcs(fs).Delims(left, right)
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

		var tmpl *template.Template
		if t == nil {
			t = template.New(name).Funcs(fs).Delims(left, right)
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
