package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jimmyfrasche/invert"
)

func Split(RS, FS string, Stdin io.Reader) (interface{}, error) {
	rs, err := cmpl(RS) //might as well add these to the cache
	if err != nil {
		return nil, err
	}

	fs, err := cmpl(FS)
	if err != nil {
		return nil, err
	}

	stdin, err := ioutil.ReadAll(Stdin)
	if err != nil {
		return nil, err
	}

	pairs := rs.FindAllIndex(stdin, -1)
	out := make([][]string, 0, len(pairs))
	for _, p := range invert.Indicies(pairs, len(stdin)) {
		s := stdin[p[0]:p[1]]

		is := fs.FindAllIndex(s, -1)
		is = invert.Indicies(is, len(s))

		row := make([]string, 0, len(is))
		for _, i := range is {
			row = append(row, string(s[i[0]:i[1]]))
		}

		out = append(out, row)
	}

	return out, nil
}

func CSV(header []string, Stdin io.Reader) (interface{}, error) {
	r := csv.NewReader(Stdin)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	if ln := len(header); ln > 0 {
		r.FieldsPerRecord = ln
	}
	recs, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, nil
	}

	if header == nil {
		header, recs = recs[0], recs[1:]
	}

	rows := []map[string]string{}
	for rn, rec := range recs {
		if h, r := len(header), len(rec); h != r {
			return nil, fmt.Errorf("%d: row len %d ≠ header len %d", rn, r, h)
		}
		row := map[string]string{}
		for i, h := range header {
			row[h] = rec[i]
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func JSON(Stdin io.Reader) (out interface{}, err error) {
	stdin, err := ioutil.ReadAll(Stdin)
	if err != nil {
		return
	}
	err = json.Unmarshal(stdin, &out)
	return
}
