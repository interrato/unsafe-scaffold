package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"
	"rsc.io/markdown"
)

var pageTemplate = template.Must(template.New("md2html").Parse(`<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width">

        <title>{{ .Title }}</title>

        <link rel="stylesheet" href="https://static.interrato.dev/styles/default.css">{{ if ne .Canonical "" }}

        <link rel="canonical" href="{{ .Canonical }}">{{ end }}{{ if eq .Canonical "https://interrato.dev/" }}
        <link rel="me" href="https://ioc.exchange/@interrato">{{ end }}{{ if ne .Description "" }}

        <meta name="description" content="{{ .Description }}">{{ end }}
    </head>
    <body>
        {{- if .Header.Title }}<h1>{{ .Title }}</h1>{{ end }}
        <main>{{ .Content }}</main>
    </body>
</html>
`))

type FrontMatter struct {
	Title       string
	Canonical   string
	Description string
	Header      struct {
		Title bool
	}
}

type Page struct {
	FrontMatter
	Content template.HTML
}

func toHTML(md []byte) template.HTML {
	var p markdown.Parser
	p.HeadingID = true
	p.Strikethrough = true
	p.Table = true
	p.SmartDot = true
	p.SmartDash = true
	p.SmartQuote = true
	doc := p.Parse(string(md))
	for _, block := range doc.Blocks {
		h, ok := block.(*markdown.Heading)
		if !ok {
			continue
		}
		if h.Level == 1 {
			errorf("level 1 headings are forbidden")
		}
	}
	return template.HTML(markdown.ToHTML(doc))
}

const usage = `Usage:
    md2html [INPUT]

Example:
    $ md2html page.md`

var name string

func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	flag.Parse()

	name = flag.Arg(0)
	if name == "" {
		flag.Usage()
		os.Exit(1)
	}
	if !strings.HasSuffix(strings.ToLower(name), ".md") {
		errorf("input file must have a .md extension")
	}

	md, err := os.ReadFile(name)
	if err != nil {
		errorf("failed to read input file: %v", err)
	}

	if bytes.Count(md, []byte("---")) < 2 {
		errorf("input file must contain a YAML front matter enclosed in '---' lines")
	}

	_, md, _ = bytes.Cut(md, []byte("---"))
	frontMatter, md, _ := bytes.Cut(md, []byte("---"))
	bytes.TrimSpace(md)

	var fm FrontMatter
	if err := yaml.Unmarshal(frontMatter, &fm); err != nil {
		errorf("failed to parse front matter: %v", err)
	}

	if fm.Title == "" {
		errorf("front matter must contain a title field")
	}
	if fm.Canonical == "" {
		warningf("front matter does not contain a canonical field")
	}

	f, err := os.Create(strings.TrimSuffix(name, ".md") + ".html")
	if err != nil {
		errorf("%s:failed to open output file: %v", err)
	}
	defer f.Close()

	page := Page{
		FrontMatter: fm,
		Content:     toHTML(md),
	}
	if err := pageTemplate.Execute(f, page); err != nil {
		errorf("failed to execute html template: %v", err)
	}
}

var l = log.New(os.Stderr, "", 0)

func warningf(format string, v ...any) {
	l.Printf("md2html("+name+"): warning: "+format, v...)
}

func errorf(format string, v ...any) {
	l.Fatalf("md2html("+name+"): error: "+format, v...)
}
