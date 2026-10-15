package main

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jub0bs/cors"
	"golang.org/x/mod/semver"
)

func main() {
	h := GlobalHandler()
	s := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("interrato.dev/.well-known/atproto-did", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("did:plc:4tzafowrls56qh7ap34c7w5x"))
	})

	mux.Handle("static.interrato.dev/styles/{style}", StaticHandler())
	mux.Handle("static.interrato.dev/fonts/{font}", WithCORS(FontHandler()))

	mux.Handle("interrato.dev/{$}", HTMLHandler("index.html"))
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
		w.Header().Set("Content-Security-Policy", "default-src 'none'; "+
			"frame-ancestors 'none'; "+
			"require-trusted-types-for 'script'; "+
			"connect-src 'self' https://static.interrato.dev; "+
			"script-src 'self' https://static.interrato.dev; "+
			"style-src 'self' https://static.interrato.dev; "+
			"font-src 'self' https://static.interrato.dev")

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

//go:embed static
var staticContent embed.FS

func StaticHandler() http.Handler {
	content, err := fs.Sub(staticContent, "static")
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServerFS(content)
}

func WithCORS(h http.Handler) http.Handler {
	cors, err := cors.NewMiddleware(cors.Config{
		Origins: []string{"https://interrato.dev", "https://*.interrato.dev"},
	})
	if err != nil {
		log.Fatal(err)
	}
	return cors.Wrap(h)
}

func FontHandler() http.Handler {
	static := StaticHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		static.ServeHTTP(w, r)
	})
}

//go:embed *.html
var htmlContent embed.FS

func HTMLHandler(name string) http.Handler {
	content, err := htmlContent.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := bytes.Clone(content)
		for _, tag := range []string{"script", "style"} {
			for bytes.Contains(response, []byte(`<`+tag+`>`)) {
				nonce := generateNonce()
				response = bytes.Replace(
					response,
					[]byte(`<`+tag+`>`),
					[]byte(`<`+tag+` nonce="`+nonce+`">`),
					1, // each one gets a different nonce
				)
				appendToCSP(w, tag+"-src", "'nonce-"+nonce+"'")
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(response)
	})
}

func generateNonce() string {
	nonce := make([]byte, 16)
	rand.Read(nonce)
	return hex.EncodeToString(nonce)
}

func appendToCSP(w http.ResponseWriter, directive, source string) {
	header := w.Header().Get("Content-Security-Policy")
	var b strings.Builder
	b.Grow(len(header) + 1 + len(source)) // current header + space + new source expression (best case)
	found := false
	for {
		policy, rest, more := strings.Cut(header, ";")
		policy = strings.TrimSpace(policy)
		if policy != "" {
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			b.WriteString(policy)
			name, _, _ := strings.Cut(policy, " ")
			if !found && strings.EqualFold(name, directive) {
				b.WriteByte(' ')
				b.WriteString(source)
				found = true
			}
		}
		if !more {
			break
		}
		header = rest
	}
	if !found {
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteString(directive)
		b.WriteByte(' ')
		b.WriteString(source)
	}
	w.Header().Set("Content-Security-Policy", b.String())
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
