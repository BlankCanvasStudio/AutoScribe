all: \
	build/autoscribe

install: \
	build/autoscribe \
	/etc/autoscribe/autoscribe.conf \
	/usr/local/bin/autoscribe

CMD_SOURCES := $(shell find cmd -name '*.go')
PKG_SOURCES := $(shell find pkg -name '*.go')

build/autoscribe: $(CMD_SOURCES) $(PKG_SOURCES)
	mkdir -p build
	go build -o $@ cmd/main.go

/etc/autoscribe/autoscribe.conf: 
	[ -f $@ ] || (mkdir -p /etc/autoscribe && cp dist/autoscribe.conf /etc/bitdrift/autoscribe.conf)

/usr/local/bin/autoscribe: build/autoscribe
	cp build/autoscribe $@

