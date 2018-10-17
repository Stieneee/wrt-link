TOOLCHAIN=/home/stieneee/toolchains/toolchain-mips_mips32_gcc-8.2.0_musl

CC=$(TOOLCHAIN)/bin/mips-openwrt-linux-gcc
STRIP=$(TOOLCHAIN)/bin/mips-openwrt-linux-musl-strip

EXTLDFLAGS=-static
LDFLAGS=-linkmode external -extldflags "$(EXTLDFLAGS)"

main: main.go Makefile
	# GOOS=linux GOARCH=mips GOARM=5 CC=$(CC) go build --ldflags='$(LDFLAGS)' -o main main.go
	GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=1 go build main.go
	$(STRIP) main

push:
	scp main ddwrt:/tmp/main

clean:
	rm -f main 
