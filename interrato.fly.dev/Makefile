all: $(patsubst %.md,%.html,$(wildcard *.md))

%.html: %.md md2html/*
	go run ./md2html $<

dev:
	watchexec --restart --exts go,html --ignore 'md2html/**' -- HTTP_HOST=interrato.dev go run .

.PHONY: all dev
