package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	Left       = flag.String("left", "{{", "set left template delimiter")
	Right      = flag.String("right", "}}", "set right template delimiter")
	Template   = flag.String("template", "", "which template to invoke, otherwise first listed")
	Expression = flag.String("e", "", "expression to use as main template")
	Html       = flag.Bool("html", false, "use html-aware automatic escaping against code injection")

	FieldSeparator  = flag.String("F", "[ \t]+", "field separator, RE2 regexp")
	RecordSeparator = flag.String("R", "\n+", "record separator, RE2 regexp")
	LinePattern     = flag.String("L", "", "line pattern, RE2 regexp")

	Json    = flag.Bool("json", false, "treat input as JSON")
	Csv     = flag.Bool("csv", false, "treat input as CSV")
	NoStdin = flag.Bool("no-stdin", false, "do not read stdin")
	Header  = flag.String("csv-header", "", "specify a header for the CSV, instead of the first row. -csv is assumed if -csv-header is used.")
)

//Usage: %name %flags template-files*
func main() {
	log.SetFlags(0)

	flag.Parse()
	args := flag.Args()

	var which string

	//parse templates

	//If -e used, use as main template, even if other templates specified.
	//If template(s) specified use first as main unless specified by flag.
	//otherwise no template
	if *Expression != "" {
		if *Template != "" {
			log.Fatalln("-template is mutually exclusive with -e")
		}
	} else if len(args) > 0 {
		which = args[0]
		if *Template != "" {
			which = *Template
		}
	} else {
		log.Fatalln("No template(s) specified")
	}
	tmpl, err := Parse(*Html, *Expression, *Left, *Right, funcs, args...)
	if err != nil {
		log.Fatalln(err)
	}

	//parse input
	var stdin interface{}
	if *Csv || *Header != "" {
		var hdr []string
		if *Header != "" {
			hdr = strings.Split(*Header, ",")
		}
		stdin, err = CSV(hdr, os.Stdin)
	} else if *Json {
		stdin, err = JSON(os.Stdin)
	} else if *LinePattern != "" {
		stdin, err = SubmatchSplit(*RecordSeparator, *LinePattern, os.Stdin)
	} else if !*NoStdin {
		stdin, err = Split(*RecordSeparator, *FieldSeparator, os.Stdin)
	}
	if err != nil {
		log.Fatalln(err)
	}

	if err = tmpl.ExecuteTemplate(os.Stdout, which, stdin); err != nil {
		log.Fatalln(err)
	}
}
