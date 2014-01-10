package main

import (
	"io"

	htmltemplate "html/template"
	texttemplate "text/template"
)

type template interface {
	New(string) template
	Parse(string) (template, error)
	ExecuteTemplate(io.Writer, string, interface{}) error
}

type textTemplate struct {
	t *texttemplate.Template
}

type htmlTemplate struct {
	t *htmltemplate.Template
}

var htmlOrText = map[bool]func(n, l, r string, f map[string]interface{}) template{
	false: newtext,
	true:  newhtml,
}

func newtext(name, left, right string, funcs map[string]interface{}) template {
	return &textTemplate{texttemplate.New(name).Delims(left, right).Funcs(funcs)}

}

func newhtml(name, left, right string, funcs map[string]interface{}) template {
	return &htmlTemplate{htmltemplate.New(name).Delims(left, right).Funcs(funcs)}
}

func (t *textTemplate) New(nm string) template {
	return &textTemplate{t.t.New(nm)}
}

func (t *textTemplate) Parse(s string) (template, error) {
	x, err := t.t.Parse(s)
	if err != nil {
		return nil, err
	}
	return &textTemplate{x}, nil
}

func (t *textTemplate) ExecuteTemplate(w io.Writer, which string, data interface{}) error {
	return t.t.ExecuteTemplate(w, which, data)
}

func (t *htmlTemplate) New(nm string) template {
	return &htmlTemplate{t.t.New(nm)}
}

func (t *htmlTemplate) Parse(s string) (template, error) {
	x, err := t.t.Parse(s)
	if err != nil {
		return nil, err
	}
	return &htmlTemplate{x}, nil
}

func (t *htmlTemplate) ExecuteTemplate(w io.Writer, which string, data interface{}) error {
	return t.t.ExecuteTemplate(w, which, data)
}
