package main

import (
	"flag"
	"log"
	"os"
)

const (
	RS = "\n+"
	FS = `\s+`
)

var (
	Left  = flag.String("left", "{{", "set left template delimiter")
	Right = flag.String("right", "}}", "set right template delimiter")

	Template   = flag.String("template", "", "which template to invoke, otherwise first listed")
	Expression = flag.String("e", "", "expression to use as main template")

	Html = flag.Bool("html", false, "use html-aware automatic escaping against code injection")

	RecordSeparator = flag.String("R", RS, "record separator")
	FieldSeparator  = flag.String("F", FS, "field separator")
	LinePattern     = flag.String("L", "", "line pattern, regex must contain capture groups")

	Json    = flag.Bool("json", false, "treat input as JSON")
	Csv     = flag.Bool("csv", false, "treat input as CSV")
	NoStdin = flag.Bool("no-stdin", false, "do not read stdin")

	Header = flag.String("header", "", "specify a header as a comma-separated list")
)

//Usage: %name %flags template-files*
func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		log.Printf("Usage: %s [-json|-csv|-no-stdin] -html -left=delim -right=delim\n", os.Args[0])
		p := log.Println
		p("\t[-e=template|-template=name] -R=RE [-F=RE|-L=RE]")
		p("\t-header=headerspec template-files*")

		p(" Template control:")
		p("  -left delim:    set the left delimiter in templates")
		p("  -right delim:   set the right delimiter in templates")
		p("  -html:          use html-aware autoescaping")
		p(" Template selection:")
		p("  -e template:    specifiy main template as string")
		p("  -template file: say which of the template files is the main template")
		p(" Input handling")
		p("  -json:          parse input as JSON")
		p("  -csv:           parse input as CSV")
		p("  -no-stdin:      do not read stdin")
		p("  -R regex:       record separator, defaults to \"\\n+\"")
		p("  -F regex:       field separator, defaults to \"\\s+\"")
		p("  -L regex:       line-matching pattern")
		p("  -header list:   comma-separated list of field names")

		p("-e and -template are mutually exclusive")
		p("Only one of -json, -csv, -no-stdin, -F, or -L can be specified")
		p("-header can only be used with -csv, -F, or -L")
		p("-R can only be used with -F or -L")

		os.Exit(2)
	}
	flag.Parse()

	//validate arguments
	fail := false
	if *Expression != "" && *Template != "" {
		fail = true
	}
	if multiple(*Csv, *Json, *NoStdin) {
		fail = true
	}
	notregex := oneOf(*Csv, *Json, *NoStdin)
	if notregex && *RecordSeparator != RS {
		fail = true
	}
	if notregex && *FieldSeparator != FS {
		fail = true
	}
	if notregex && *LinePattern != "" {
		fail = true
	}
	if (*Json || *NoStdin) && *Header != "" {
		fail = true
	}
	if *FieldSeparator != FS && *LinePattern != "" {
		fail = true
	}
	if fail {
		log.Println("Invalid combination of flags")
		flag.Usage()
		os.Exit(2)
	}
	args := flag.Args()

	var which string

	//parse templates

	//If -e used, use as main template, even if other templates specified.
	//If template(s) specified use first as main unless specified by flag.
	//otherwise no template
	if len(args) > 0 {
		which = args[0]
		if *Template != "" {
			which = *Template
		}
	} else if *Expression == "" {
		log.Fatalln("No template(s) specified")
	}
	tmpl, err := Parse(*Html, *Expression, *Left, *Right, funcs, args...)
	if err != nil {
		log.Fatalln(err)
	}

	hdr := splitHeader(*Header)

	//parse input
	var stdin interface{}
	if *Csv {
		stdin, err = CSV(hdr, os.Stdin)
	} else if *Json {
		stdin, err = JSON(os.Stdin)
	} else if *LinePattern != "" {
		stdin, err = SubmatchSplit(hdr, *RecordSeparator, *LinePattern, os.Stdin)
	} else if !*NoStdin {
		stdin, err = Split(hdr, *RecordSeparator, *FieldSeparator, os.Stdin)
	}
	if err != nil {
		log.Fatalln(err)
	}

	//run program
	if err = tmpl.ExecuteTemplate(os.Stdout, which, stdin); err != nil {
		log.Fatalln(err)
	}
}
