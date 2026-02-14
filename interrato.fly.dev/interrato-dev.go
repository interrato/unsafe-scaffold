package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"filippo.io/age"
)

//go:embed interrato.dev
var interratoDevContent embed.FS

var goGetTmpl = template.Must(template.New("go-get").Parse(`
{{ $repo := or .GitRepo (printf "https://github.com/BuriedInTheGround/%s" .Name) }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta name="go-import" content="interrato.dev/{{ .Name }} git {{ $repo }}">
    <meta http-equiv="refresh" content="0;url={{ or .Redirect $repo }}">
</head>
<body>
    You will be redirected to the <a href="{{ or .Redirect $repo }}">project page</a>...
</body>
</html>
`))

type goGetHandler struct {
	Name     string
	GitRepo  string
	Redirect string
}

func (h goGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	goGetTmpl.Execute(w, h)
}

//go:embed apprendimento/*.html
var apprendimentoContent embed.FS

func apprendimentoHandler(name string) http.Handler {
	content, err := apprendimentoContent.ReadFile(name)
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})
}

var pages = []struct {
	Title        string
	Description  string
	Path         string
	TemplateName string
	Data         any
}{
	{
		Title:        "", // no title for the homepage
		Description:  "Computer Engineer, Cybersecurity at the University of Padua.",
		Path:         "/",
		TemplateName: "home.html",
	},
	{
		Title:        "Notes",
		Description:  "Index of Simone's notes.",
		Path:         "/notes/",
		TemplateName: "notes.html",
		Data:         notes,
	},
	{
		Title:        "Resume",
		Description:  "Simone's resume.",
		Path:         "/resume/",
		TemplateName: "resume.html",
	},
}

var notes = []struct {
	Title       string
	Description string
	Path        string // left empty, constructed later as /notes/{slug}/
	Slug        string
	Date        string
	Protected   bool
}{
	{
		Title:       "Perfectionism slows me down, but it is also useful",
		Description: "I am going to relearn to make mistakes, but I won't stop striving for perfection because it guides me to produce quality results.",
		Slug:        "perfectionism",
		Date:        "2024-07-18",
	},
	{
		Title:       "Taisha-ryū",
		Description: "Personal notes about Taisha-ryū.",
		Slug:        "taisha",
		Date:        "2024-01-15",
		Protected:   true,
	},
}

func interratoDev(mux *http.ServeMux) {
	mux.HandleFunc("www.interrato.dev/", func(w http.ResponseWriter, r *http.Request) {
		u := &url.URL{
			Scheme:   "https",
			Host:     "interrato.dev",
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	})

	content, err := fs.Sub(interratoDevContent, "interrato.dev")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("interrato.dev/static/", http.FileServer(http.FS(content)))

	mux.Handle("interrato.dev/apprendimento", apprendimentoHandler("apprendimento/index.html"))
	for i := range 9 {
		path := fmt.Sprintf("apprendimento/lezione%d", i+1)
		mux.Handle("interrato.dev/"+path, apprendimentoHandler(path+".html"))
	}
	mux.Handle(
		"interrato.dev/apprendimento/cheatsheet",
		apprendimentoHandler("apprendimento/cheatsheet.html"),
	)

	funcs := template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}

	for _, page := range pages {
		mux.HandleFunc("interrato.dev"+page.Path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != page.Path {
				http.NotFound(w, r)
				return
			}
			tmpl, err := template.New("base.html").Funcs(funcs).ParseFS(
				content,
				"templates/base.html",
				"templates/"+page.TemplateName,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tmpl.Execute(w, page)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}

	for _, note := range notes {
		note.Path = "/notes/" + note.Slug + "/"
		mux.HandleFunc("interrato.dev"+note.Path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != note.Path {
				http.NotFound(w, r)
				return
			}
			filename := note.Slug
			if note.Protected {
				filename = "protected"
			}
			tmpl, err := template.New("base.html").Funcs(funcs).ParseFS(
				content,
				"templates/base.html",
				"templates/notes/base.html",
				"templates/notes/"+filename+".html",
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tmpl.Execute(w, note)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		mux.HandleFunc("POST interrato.dev"+note.Path, func(w http.ResponseWriter, r *http.Request) {
			identity, err := age.NewScryptIdentity(r.PostFormValue("passphrase"))
			identity.SetMaxWorkFactor(16) // limit memory usage to ~64 MiB
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f, err := content.Open("templates/notes/" + note.Slug + ".html.age")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			out, err := age.Decrypt(f, identity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl, err := template.New("base.html").Funcs(funcs).ParseFS(
				content,
				"templates/base.html",
				"templates/notes/base.html",
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			outBytes, err := io.ReadAll(out)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl, err = tmpl.Parse(string(outBytes))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tmpl.Execute(w, note)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}

	// go get handlers
	mux.Handle("interrato.dev/carbonize/", goGetHandler{
		Name: "carbonize",
	})
	mux.Handle("interrato.dev/fine/", goGetHandler{
		Name: "fine",
	})
	mux.Handle("interrato.dev/olaf/", goGetHandler{
		Name: "olaf",
	})
	mux.Handle("interrato.dev/pigowa/", goGetHandler{
		Name: "pigowa",
	})
}
