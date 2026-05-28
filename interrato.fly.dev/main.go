package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
	mux.Handle("www.perpetuatheme.com/", HostRedirectHandler("interrato.dev", http.StatusMovedPermanently))

	mux.Handle("perpetuatheme.com/{$}", http.RedirectHandler("https://github.com/perpetuatheme", http.StatusFound))

	// CSP style-src hashes
	var styles []string

	mux.Handle("interrato.dev/{$}", StaticHandler())
	styles = append(styles, "sha256-oVjEMD7V6zajFUyDJYhT6JvHQ5kOapiRt6TgQ4QRiFE=")
	mux.Handle("interrato.dev/static/fonts/", StaticHandler())
	mux.Handle("interrato.dev/static/pdf/", StaticHandler())

	mux.Handle("interrato.dev/apprendimento/", HTMLHandler("apprendimento.html"))
	mux.Handle("interrato.dev/infosec/", HTMLHandler("infosec.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
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
