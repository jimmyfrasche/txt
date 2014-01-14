#txt
Command txt is a templating language for shell programming.

Download:
```shell
go get github.com/jimmyfrasche/txt
```

* * *
Command txt is a templating language for shell programming.

##Input
The input to the template comes from stdin.
It is parsed in one of five ways.

The default is to split stdin into records and fields, using the -R and -F
flags respectively, similar to awk(1), and dot is set to a list of records
(see below).
If header is set, dot is a list of maps with the specified names as keys.

If the -L flag is specified stdin is broken into records as with the default,
but the fields are defined by the capture groups of the regular expression
-L.
Records that do not match -L are skipped.
If -L contains named capture groups each record is a dictionary of only
the named captures' values for that record.
Otherwise, dot is a list of records (see below) of the capture groups' values.
If header is set, dot is a list of maps with the specified names as keys,
overriding any names from capture groups.

If the -csv flag is specified, stdin is treated as a CSV file, as recognized
by the encoding/csv package.
If the -header flag is not specified, the first record is used as the header.
Dot is set to a list of maps, with the header for each column as the key.

If the -json flag is specified, stdin is treated as JSON.
Dot is set as the decoded JSON.

If the -no-stdin flag is specified, stdin is not read.
Dot is not set.

##Records
When using -F or -L without a header, or in the case of -L without named
capture groups, dot is a list of records.

Each record has two fields, Fields and Line.
Line is the complete unaltered input of that record.
Fields are the values of each field in that record.
If dot is a record

```
{{.}}
```

is the same as

```
{{.Line}}
```

Records have a method F that takes an integer n and returns the nth field
if it exists and the empty string otherwise.
If n is negative it returns the (n-1)th field from the end.

If n is positive and the nth field exists, then

```
{{.F n}}
```

is equivalent to

```
{{index . n}}
```

##Templates
The templating language is documented at http://golang.org/pkg/text/template
with the single difference that if the first line at the top of the file
begins with #! that line is skipped.
If the -html flag is used, escaping functions are automatically added to all
outputs based on context.

Any command line arguments after the flags are treated as filenames
of templates.
The templates are named after the basename of the respective filename.
The first file listed is the main template, unless the -template flag
specifies otherwise.
If the -e flag is used to define an inline template, it is always the main
template, and the -template flag is illegal.

##Regular Expressions
All regular expressions are RE2 regular expression with the Perl syntax and
semantics.
The syntax is documented at
http://golang.org/pkg/regexp/syntax/#hdr-Syntax

##Functions
Built in functions are documented at
http://golang.org/pkg/text/template#hdr-Functions

The following additional functions are defined:

```
slice what start stop
	slices what from start to stop.
	stop is optional. what must be a list or string.
	start or stop may be negative. If start or stop exceed the bounds
	of what, the largest slice of what that exists is returned.

nl string
	append a newline to the end of string, if it does not end in newline.

readCSV headerspec filename
	headerspec is a comma-separated list of headers or "" to use the headers
	in filename.
	Dot is set to the contents of the CSV file as with -csv.
	If the file cannot be opened or its contents are malformed, execution
	stops.

readJSON filename
	Read the JSON encoded file into dot or halt execution if decoding fails
	or the file cannot be opened.
	Dot is set to the contents of the JSON file as with -json.

readLine header FS LP filename
	Read filename with line pattern splitting as specified by the RS and
	LP regular expressions, and an optional header header.
	If header is not "", the names in header will be used as the field names.
	If RS or LP are "", the respective value of -R or -L is used.

read header RS FS filename
	Read filename with the default record and file splitting as specified
	by the RS and FS regular expressions, and optional header header.
	If header is not "", the names in header will be used as the field names.
	If RS or FS are "", the respective value of -R or -F is used.

readFile filename
	Read filename completely as a single string.
	Execution halts if the file cannot be read.

quoteCSV string
	Apply the appropriate CSV quoting rules to string.

toJSON what
	Encode what as JSON. Execution halts if
	http://golang.org/pkg/encoding/json/#Marshal errors.

equalFold string-one string-two
	Reports whether the UTF-8 encoded string-one and string-two are equal
	under Unicode case-folding.

fields string
	Split string around whitespace.

join separator strings
	Join the list in strings by the string separator.

lower string
	Lowercase string.

upper string
	Uppercase string.

title string
	Titlecase string.

trim cutset string
	Return string with all leading and trailing runes in cutset removed.

trimLeft cutset string
	Return string with all leading runes in cutset removed.

trimRight cutset string
	Return string with all trailing runes in cutset removed.

trimPrefix prefix string
	Return string with prefix removed.

trimSuffix suffix string
	Return string with suffix removed.

trimSpace string
	Return string with all leading and trailing whitespace removed.

quoteGo string
	Return string quoted as a Go string literal. Escapes non-printable
	runes. Should work for most languages that accept UTF-8 source.

quoteGoASCII string
	As quoteGo except any non-ASCII runes are escaped to hexcodes.
	Should work for most languages.

match pattern string
	Return whether string matches the regex in pattern.
	Execution halts if pattern is not a valid regular expression.

find pattern string
	Returns all substrings of string that match pattern.
	Execution halts if pattern is not a valid regular expression.

replace pattern spec string
	Replace all substrings in string matching pattern by spec.
	Execution halts if pattern is not a valid regular expression.

split pattern string
	Split string into a list of substrings separated by pattern.
	Execution halts if pattern is not a valid regular expression.

env key
	Returns the environment variable key or "".

exec name args*
	Execute command name with args. Stdin is nil.
	Stderr shares the stderr of txt(1).
	Stdout is returned as a string.

pipe name args* input
	Execute command name with args with input as stdin.
	Otherwise, like exec.
```



* * *
Automatically generated by [autoreadme](https://github.com/jimmyfrasche/autoreadme) on 2014.01.13
