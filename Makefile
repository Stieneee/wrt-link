LDFLAGS=-s -w -extldflags -static

GOFILES=main.go ipconntrack.go iptable.go reporter.go raven.go jwt.go

wrt-link_linux_mips: $(GOFILES)
	GOOS=linux GOARCH=mips GOMIPS=softfloat go build -o $@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute $@

wrt-link_linux_mipsle: $(GOFILES)
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o $@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute $@

wrt-link_linux_arm: $(GOFILES)
	GOOS=linux GOARCH=arm go build -o $@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute $@

wrt-link_linux_arm64: $(GOFILES)
	GOOS=linux GOARCH=arm64 go build -o $@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute $@

wrt-link_linux_amd64: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -o $@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute $@

gox:
	gox -os="linux" -ldflags="$(LDFLAGS)"
	upx --best --ultra-brute $(wildcard wrt-link_*) || true

push:
	scp init-wrt-link.sh ddwrt:/tmp/init-wrt-link.sh
	scp wrt-link_linux_mips ddwrt:/tmp/wrt-link

clean:
	rm -f wrt-link*

.PHONY: gox push clean