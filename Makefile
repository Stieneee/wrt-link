
DATE := $(shell date -u --iso-8601=minutes)
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)

LDFLAGS=-s -w -extldflags -static -X 'main.BuildTime=$(DATE)' -X 'main.BuildVersion=$(VERSION)'

GOFILES=main.go ipconntrack.go iptable.go reporter.go raven.go jwt.go

wrt-link_linux_mips: $(GOFILES)
	echo "$(LDFLAGS)"
	GOOS=linux GOARCH=mips GOMIPS=softfloat go build -o build/$@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute build/$@

wrt-link_linux_mipsle: $(GOFILES)
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o build/$@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute build/$@

wrt-link_linux_arm: $(GOFILES)
	GOOS=linux GOARCH=arm go build -o build/$@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute build/$@

wrt-link_linux_arm64: $(GOFILES)
	GOOS=linux GOARCH=arm64 go build -o build/$@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute build/$@

wrt-link_linux_amd64: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -o build/$@ --ldflags='$(LDFLAGS)' $(GOFILES)
	upx --best --ultra-brute build/$@

gox:
	gox -os="linux" -ldflags="$(LDFLAGS)" -output "build/wrt-link_{{.OS}}_{{.Arch}}"
	-upx -q --best --ultra-brute build/wrt-link_linux_* 
	cp init-wrt-link.sh build/
	sed -i "s/VERSIONPLACEHOLDER/$(VERSION)/g" build/init-wrt-link.sh 
	# VERSION needs to be fixed
 
push: 
	scp init-wrt-link.sh ddwrt:/tmp/init-wrt-link.sh
	scp build/wrt-link_linux_mips ddwrt:/tmp/wrt-link

clean:
	rm -rf build

.PHONY: gox gox-quiet push clean