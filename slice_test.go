package main

import (
	"fmt"
	"sort"
	"testing"
)

func fields(ret interface{}) (out [][]string, err error) {
	recs, ok := ret.([]*record)
	if !ok {
		return nil, fmt.Errorf("wrong return type, expected []*record, got %T", ret)
	}
	for _, rec := range recs {
		out = append(out, rec.Fields)
	}
	return
}

func listEquals(r int, a, b []string) error {
	if la, lb := len(a), len(b); la != lb {
		return fmt.Errorf("record %d: wrong number of fields: %d ≠ %d", r, la, lb)
	}
	for f := range a {
		if a[f] != b[f] {
			return fmt.Errorf("record %d, field %d: %#v ≠ %#v", r, f, a[f], b[f])
		}
	}
	return nil
}

func keyValues(m map[string]string) (keys, values []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values = make([]string, len(keys))
	for i, k := range keys {
		values[i] = m[k]
	}
	return
}

func listListEquals(a, b [][]string) error {
	if la, lb := len(a), len(b); la != lb {
		return fmt.Errorf("wrong number of records: %d ≠ %d", la, lb)
	}
	for i := range a {
		if err := listEquals(i, a[i], b[i]); err != nil {
			return err
		}
	}
	return nil
}

func listMapEquals(a, b []map[string]string) error {
	if la, lb := len(a), len(b); la != lb {
		return fmt.Errorf("wrong number of records: %d ≠ %d", la, lb)
	}
	for i := range a {
		ai, bi := a[i], b[i]
		if la, lb := len(ai), len(bi); la != lb {
			return fmt.Errorf("record %d: wrong number of fields: %d ≠ %d", i, la, lb)
		}
		aki, avi := keyValues(ai)
		bki, bvi := keyValues(bi)
		if err := listEquals(i, aki, bki); err != nil {
			return fmt.Errorf("key space: %s", err)
		}
		if err := listEquals(i, avi, bvi); err != nil {
			return err
		}
	}
	return nil
}

func failIf(t *testing.T, testno int, err error) {
	if err != nil {
		t.Errorf("test case %d: %s", testno, err)
	}
}

//TEST SPLIT

var splitNoHeaderTests = []struct {
	corpus string
	out    [][]string
}{
	{
		corpus: "a b c\nd e f\ng h i",
		out: [][]string{
			{"a", "b", "c"},
			{"d", "e", "f"},
			{"g", "h", "i"},
		},
	},
	{
		corpus: "a b c\n \ng h i j",
		out: [][]string{
			{"a", "b", "c"},
			{},
			{"g", "h", "i", "j"},
		},
	},
}

func TestSplitNoHeader(t *testing.T) {
	for i, v := range splitNoHeaderTests {
		ret, err := Split(nil, RS, FS, rdr(v.corpus))
		failIf(t, i, err)
		recs, err := fields(ret)
		failIf(t, i, err)
		failIf(t, i, listListEquals(v.out, recs))
	}
}

var splitHeaderTests = []struct {
	corpus string
	header []string
	out    []map[string]string
}{
	{
		corpus: "1 2 3\n 4 5 6 \n7 8 9\n",
		header: []string{"a", "b", "c"},
		out: []map[string]string{
			{"a": "1", "b": "2", "c": "3"},
			{"a": "4", "b": "5", "c": "6"},
			{"a": "7", "b": "8", "c": "9"},
		},
	},
	{
		corpus: "1 2 3\n 4 5 6 \n7 8 9 10\n",
		header: []string{"a", "b", "c"},
		out: []map[string]string{
			{"a": "1", "b": "2", "c": "3"},
			{"a": "4", "b": "5", "c": "6"},
			{"a": "7", "b": "8", "c": "9"},
		},
	},
	{
		corpus: "1 2 3\n 4 5 6 \n7 8\n",
		header: []string{"a", "b", "c"},
		out: []map[string]string{
			{"a": "1", "b": "2", "c": "3"},
			{"a": "4", "b": "5", "c": "6"},
			{"a": "7", "b": "8", "c": ""},
		},
	},
}

func TestSplitHeader(t *testing.T) {
	for i, v := range splitHeaderTests {
		ret, err := Split(v.header, RS, FS, rdr(v.corpus))
		failIf(t, i, err)
		failIf(t, i, listMapEquals(v.out, ret.([]map[string]string)))
	}
}

//TEST SUBMATCH

func TestSubmatchSplitBadPattern(t *testing.T) {
	_, err := SubmatchSplit(nil, RS, "\t+", rdr(""))
	if err == nil || err.Error() != "submatch splitting requires a regexp with submatches" {
		t.Errorf("wrong error message returned: %v", err)
	}
}

var submatchNoHeaderNoNamesTests = []struct {
	corpus, LP string
	out        [][]string
}{
	{
		corpus: " key = value ; comment ",
		LP:     `([^\s;=]+)\s*=\s*([^;]+)`,
		out:    [][]string{{"key", "value "}},
	},
}

func TestSubmatchSplitNoHeaderNoNames(t *testing.T) {
	for i, v := range submatchNoHeaderNoNamesTests {
		ret, err := SubmatchSplit(nil, RS, v.LP, rdr(v.corpus))
		failIf(t, i, err)
		recs, err := fields(ret)
		t.Log(recs)
		failIf(t, i, err)
		failIf(t, i, listListEquals(v.out, recs))
	}
}

type submatchHeader struct {
	corpus string
	header []string
	LP     string
	out    []map[string]string
}

var submatchNoHeaderNamesTests = []submatchHeader{
	{
		corpus: " key = value ; comment ",
		LP:     `(?P<k>[^\s;=]+)\s*=\s*(?P<v>[^;]+)`,
		out:    []map[string]string{{"k": "key", "v": "value "}},
	},
}

func TestSubmatchSplitNoHeaderNames(t *testing.T) {
	for i, v := range submatchNoHeaderNamesTests {
		ret, err := SubmatchSplit(v.header, RS, v.LP, rdr(v.corpus))
		failIf(t, i, err)
		failIf(t, i, listMapEquals(v.out, ret.([]map[string]string)))
	}
}

var submatchHeaderNoNamesTests = []submatchHeader{
	{
		corpus: " key = value ; comment ",
		LP:     `([^\s;=]+)\s*=\s*([^;]+)`,
		header: []string{"k", "v"},
		out:    []map[string]string{{"k": "key", "v": "value "}},
	},
}

func TestSubmatchSplitHeaderNoNames(t *testing.T) {
	for i, v := range submatchHeaderNoNamesTests {
		ret, err := SubmatchSplit(v.header, RS, v.LP, rdr(v.corpus))
		failIf(t, i, err)
		failIf(t, i, listMapEquals(v.out, ret.([]map[string]string)))
	}
}

var submatchHeaderNamesTests = []submatchHeader{
	{
		corpus: " key = value ; comment ",
		LP:     `(?P<key>[^\s;=]+)\s*=\s*(?P<value>[^;]+)`,
		header: []string{"k", "v"},
		out:    []map[string]string{{"k": "key", "v": "value "}},
	},
}

func TestSubmatchSplitHeaderNames(t *testing.T) {
	for i, v := range submatchHeaderNamesTests {
		ret, err := SubmatchSplit(v.header, RS, v.LP, rdr(v.corpus))
		failIf(t, i, err)
		failIf(t, i, listMapEquals(v.out, ret.([]map[string]string)))
	}
}
