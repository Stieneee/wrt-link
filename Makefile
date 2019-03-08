TOOLCHAIN=/usr/local/ddwrt/toolchain-mips_mips32_gcc-8.2.0_musl

LDFLAGS=-extldflags "-static"

wrt-link-mips: main.go ipconntrack.go iptable.go reporter.go raven.go jwt.go
	GOOS=linux GOARCH=mips GOMIPS=softfloat C=$(TOOLCHAIN)/bin/mips-openwrt-linux-gcc go build -o $@ --ldflags='$(LDFLAGS)' main.go ipconntrack.go iptable.go reporter.go raven.go jwt.go
	$(TOOLCHAIN)/bin/mips-openwrt-linux-musl-strip $@
	upx --best --ultra-brute $@

get-toolchain:
	mkdir -p /usr/local/ddwrt/
	rm -rf /usr/local/ddwrt/*
	curl  https://wrt-link.sfo2.digitaloceanspaces.com/toolchains.tar.xz -o toolchains.tar.xz 
	tar xvJf /tmp/toolchains.tar.xz -C /usr/local/ddwrt/
	rm /tmp/toolchains.tar.xz

push:
	scp wrt-link-mips ddwrt:/tmp/wrt-link

clean:
	rm -f wrt-link-*
