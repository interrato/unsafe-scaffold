package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

func main() {
	h := GlobalHandler()
	s := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if host := os.Getenv("HTTP_HOST"); host != "" && r.Host == "localhost:8080" {
				r.Host = host
			}
			h.ServeHTTP(w, r)
		}),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  10 * time.Minute,
	}
	log.Fatal(s.ListenAndServe())
}

func GlobalHandler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("www.interrato.dev/", HostRedirectHandler("interrato.dev", http.StatusMovedPermanently))
	mux.Handle("www.emys-sse.org/", HostRedirectHandler("emys-sse.org", http.StatusMovedPermanently))
	mux.Handle("www.perpetuatheme.com/", HostRedirectHandler("perpetuatheme.com", http.StatusMovedPermanently))

	mux.Handle("emys-sse.org/{$}", http.RedirectHandler("https://github.com/interrato/emys", http.StatusFound))
	mux.Handle("perpetuatheme.com/{$}", http.RedirectHandler("https://github.com/perpetuatheme", http.StatusFound))

	// CSP style-src hashes
	var styles []string

	mux.Handle("interrato.dev/{$}", StaticHandler())
	styles = append(styles, "sha256-oVjEMD7V6zajFUyDJYhT6JvHQ5kOapiRt6TgQ4QRiFE=")
	mux.Handle("interrato.dev/static/fonts/", StaticHandler())
	mux.Handle("interrato.dev/static/pdf/", StaticHandler())

	mux.Handle("interrato.dev/apprendimento/", HTMLHandler("apprendimento.html"))
	mux.Handle("interrato.dev/infosec/", HTMLHandler("infosec.html"))

	interratoDEVModules := []string{"can", "carbonize", "emys", "fine", "olaf", "unsafe-scaffold"}

	mux.Handle("interrato.dev/{path...}", PkgsiteHandler("interrato.dev", interratoDEVModules))

	goGetMux := http.NewServeMux()
	for _, name := range interratoDEVModules {
		module := "interrato.dev/" + name
		goGetMux.Handle(module, GoImportHandler(module, "https://github.com/interrato/"+name))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		if r.URL.Query().Get("go-get") == "1" {
			goGetMux.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		var b strings.Builder
		for _, hash := range styles {
			b.WriteString(" '")
			b.WriteString(hash)
			b.WriteString("'")
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; style-src"+b.String())

		mux.ServeHTTP(w, r)
	})
}

func HostRedirectHandler(target string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := &url.URL{
			Scheme:   "https",
			Host:     target,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		http.Redirect(w, r, u.String(), code)
	})
}

//go:embed interrato.dev
var interratoDEVContent embed.FS

func StaticHandler() http.Handler {
	content, err := fs.Sub(interratoDEVContent, "interrato.dev")
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServerFS(content)
}

//go:embed *.html
var htmlContent embed.FS

func HTMLHandler(name string) http.Handler {
	content, err := htmlContent.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})
}

func GoImportHandler(module, repo string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><meta name="go-import" content="%s git %s">`, module, repo)
	})
}

func PkgsiteHandler(host string, modules []string) http.Handler {
	re := regexp.MustCompile(`([a-z0-9-]+)(?:[.]([A-Z][\w\d.]*))?(?:@([a-z0-9.-]+))?(?:[/]([a-z0-9/.-]*[a-z0-9/-]))?(?:[.]([A-Z][\w\d.]*))?`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("path")
		groups := re.FindStringSubmatch(path)
		if len(groups) != 6 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		mod := groups[1]
		sym := groups[2]
		ver := groups[3]
		pkg := groups[4]
		if pkg != "" {
			sym = groups[5]
		}
		if ver != "" && !semver.IsValid(ver) || !slices.Contains(modules, mod) {
			http.NotFound(w, r)
			return
		}
		u := &url.URL{
			Scheme:   "https",
			Host:     "pkg.go.dev",
			Path:     "/" + host + "/" + mod,
			RawQuery: r.URL.RawQuery,
		}
		if ver != "" {
			u.Path += "@" + ver
		}
		if pkg != "" {
			u.Path += "/" + pkg
		}
		if sym != "" {
			u.Fragment = sym
		}
		http.Redirect(w, r, u.String(), http.StatusFound)
	})
}
