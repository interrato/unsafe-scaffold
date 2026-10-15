all: $(patsubst %.md,%.html,$(wildcard *.md))

%.html: %.md md2html/*
	go run ./md2html $<

dev: hosts
	$(MAKE) --jobs=2 website proxy

website:
	watchexec --restart --exts go,html,css,js,wasm --ignore 'md2html/**' -- go run .

proxy:
	sudo caddy run # Binding to port 80 requires root privileges

hosts: clean
	echo '127.0.0.1 interrato.dev static.interrato.dev' | sudo tee --append /etc/hosts

clean:
	grep --invert-match interrato.dev /etc/hosts | sudo tee /etc/hosts

.PHONY: all dev website proxy hosts clean
