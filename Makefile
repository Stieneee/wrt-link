TOOLCHAIN=/home/stieneee/toolchains/toolchain-mips_mips32_gcc-8.2.0_musl

CC=$(TOOLCHAIN)/bin/mips-openwrt-linux-gcc
STRIP=$(TOOLCHAIN)/bin/mips-openwrt-linux-musl-strip

LDFLAGS=-s -w -extldflags "-static"

main: main.go Makefile
	# GOOS=linux GOARCH=mips GOARM=5 CC=$(CC) go build --ldflags='$(LDFLAGS)' -o main main.go
	GOOS=linux GOARCH=mips GOMIPS=softfloat CC=$(CC) go build --ldflags='$(LDFLAGS)' main.go ipconntrack.go iptable.go
	$(STRIP) main

push:
	scp main ddwrt:/tmp/main

clean:
	rm -f main 
