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

        <style>
            @font-face {
                font-display: swap;
                font-family: 'Cormorant SC';
                font-style: normal;
                font-weight: 400;
                src: url('/static/fonts/cormorant-sc-v19-latin-regular.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Cormorant SC';
                font-style: normal;
                font-weight: 700;
                src: url('/static/fonts/cormorant-sc-v19-latin-700.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Cormorant';
                font-style: normal;
                font-weight: 400;
                src: url('/static/fonts/cormorant-v24-latin-regular.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Cormorant';
                font-style: normal;
                font-weight: 700;
                src: url('/static/fonts/cormorant-v24-latin-700.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Ysabeau';
                font-style: normal;
                font-weight: 400;
                src: url('/static/fonts/ysabeau-v5-latin-regular.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Ysabeau';
                font-style: italic;
                font-weight: 400;
                src: url('/static/fonts/ysabeau-v5-latin-italic.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Ysabeau';
                font-style: normal;
                font-weight: 700;
                src: url('/static/fonts/ysabeau-v5-latin-700.woff2') format('woff2');
            }

            @font-face {
                font-display: swap;
                font-family: 'Courier Prime';
                font-style: normal;
                font-weight: 400;
                src: url('/static/fonts/courier-prime-v11-latin-regular.woff2') format('woff2');
            }

            :root {
                color-scheme: light dark;
                -webkit-font-smoothing: antialiased;
                -moz-osx-font-smoothing: grayscale;
                text-rendering: geometricPrecision;
            }

            body {
                font-family: 'Ysabeau', sans-serif;
                font-size: 1.25rem;
                max-width: 780px;
                margin-block: 80px;
                margin-inline: auto;
                padding-inline: 20px;
            }

            h1 {
                font-family: 'Cormorant SC', serif;
                text-wrap: balance;
            }

            h2, h3 {
                font-family: 'Cormorant', serif;
                text-wrap: balance;
            }

            p {
                max-width: 63ch;
            }

            code {
                font-family: 'Courier Prime', monospace;
                font-size: 0.92em;
            }

            a[rel~='external'][target='_blank']::after {
                content: ' ↗';
            }

            nav {
                display: flex;
                flex-wrap: wrap;
                column-gap: 1.2em;
                row-gap: 0.3lh;
            }

            nav > a {
                white-space: nowrap;
            }

            .stone {
                display: none;
            }
        </style>{{ if ne .Canonical "" }}

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
