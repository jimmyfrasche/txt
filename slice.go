package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jimmyfrasche/invert"
)

type record struct {
	Fields []string
	Line   string
}

func (r *record) F(n int) string {
	ln := len(r.Fields)
	if n < 0 {
		n = ln + n
	}
	if n < 0 || n >= ln {
		return ""
	}
	return r.Fields[n]
}

func (r *record) String() string {
	return r.Line
}

func SubmatchSplit(header []string, RS, LinePattern string, Stdin io.Reader) (ret interface{}, err error) {
	rs, err := cmpl(RS) //might as well add these to the cache
	if err != nil {
		return
	}

	lp, err := cmpl(LinePattern)
	if err != nil {
		return
	}

	//code loosely based on but entirely inspired by rsc's reply to
	//https://groups.google.com/forum/#!topic/golang-nuts/4LpRZDfNXIc
	if lp.NumSubexp() < 1 {
		return nil, errors.New("submatch splitting requires a regexp with submatches")
	}
	names := hdr2map(header, true)
	if len(names) == 0 {
		for i, name := range lp.SubexpNames() {
			if name != "" {
				names[name] = i
			}
		}
	}

	stdin, err := ioutil.ReadAll(Stdin)
	if err != nil {
		return
	}

	pairs := rs.FindAllIndex(stdin, -1)
	records := invert.Indicies(pairs, len(stdin))
	if len(names) > 0 {
		//if there are named submatches build the row as a map with just the named entries.
		//multiple names are by construction the value of the last name.
		out := make([]map[string]string, 0, len(records))
		for _, p := range records {
			if sms := lp.FindSubmatch(stdin[p[0]:p[1]]); sms != nil {
				row := make(map[string]string, len(names))
				for name, i := range names {
					if i < len(sms) {
						row[name] = string(sms[i])
					} else {
						row[name] = ""
					}
				}
				out = append(out, row)
			}

		}
		ret = out
	} else {
		out := make([]*record, 0, len(records))
		for _, p := range records {
			line := stdin[p[0]:p[1]]
			if sms := lp.FindSubmatch(line); sms != nil {
				row := make([]string, 0, len(sms)-1)
				for _, sm := range sms[1:] {
					row = append(row, string(sm))
				}
				out = append(out, &record{
					Fields: row,
					Line:   string(line),
				})
			}
		}
		ret = out
	}

	return
}

func Split(header []string, RS, FS string, Stdin io.Reader) (ret interface{}, err error) {
	rs, err := cmpl(RS)
	if err != nil {
		return
	}

	fs, err := cmpl(FS)
	if err != nil {
		return
	}

	stdin, err := ioutil.ReadAll(Stdin)
	if err != nil {
		return
	}

	names := hdr2map(header, false)

	pairs := rs.FindAllIndex(stdin, -1)
	records := invert.Indicies(pairs, len(stdin))
	if len(names) > 0 {
		out := make([]map[string]string, 0, len(records))
		for _, p := range records {
			s := stdin[p[0]:p[1]]

			is := fs.FindAllIndex(s, -1)
			is = invert.Indicies(is, len(s))

			row := make(map[string]string, len(names))
			for name, i := range names {
				if i < len(is) {
					ix := is[i]
					row[name] = string(s[ix[0]:ix[1]])
				} else {
					row[name] = ""
				}
			}
			out = append(out, row)
		}
		ret = out
	} else {
		out := make([]*record, 0, len(records))
		for _, p := range records {
			s := stdin[p[0]:p[1]]

			is := fs.FindAllIndex(s, -1)
			is = invert.Indicies(is, len(s))

			row := make([]string, 0, len(is))
			for _, ix := range is {
				row = append(row, string(s[ix[0]:ix[1]]))
			}

			out = append(out, &record{
				Fields: row,
				Line:   string(s),
			})
		}
		ret = out
	}

	return
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
