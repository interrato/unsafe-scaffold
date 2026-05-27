package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	h := MuxHandler()
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

func MuxHandler() http.Handler {
	mux := http.NewServeMux()

	// CSP style-src hashes
	var styles []string

	mux.Handle("interrato.dev/{$}", StaticHandler())
	styles = append(styles, "sha256-Vi7t4iKt3vd3m6kSdzNhTzxGEoHplRNXq56Bq4i6+Yo=")
	mux.Handle("interrato.dev/static/pdf/", StaticHandler())

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

//go:embed interrato.dev
var interratoDEVContent embed.FS

func StaticHandler() http.Handler {
	content, err := fs.Sub(interratoDEVContent, "interrato.dev")
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServerFS(content)
}
