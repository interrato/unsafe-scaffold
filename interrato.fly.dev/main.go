package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
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

	mux.Handle("interrato.dev/{$}", StaticHandler())
	mux.Handle("interrato.dev/static/", StaticHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
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
