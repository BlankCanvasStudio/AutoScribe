all: \
	build/autoscribe

install: \
	build/autoscribe \
	/etc/autoscribe/conf.yml \
	/etc/autoscribe/prompts \
	/usr/local/bin/autoscribe

CMD_SOURCES := $(shell find cmd -name '*.go')
PKG_SOURCES := $(shell find pkg -name '*.go')

build/autoscribe: $(CMD_SOURCES) $(PKG_SOURCES)
	mkdir -p build
	go build -o $@ cmd/main.go

/etc/autoscribe/conf.yml:
	[ -f $@ ] || (mkdir -p /etc/autoscribe && cp dist/conf.yml $@)

/usr/local/bin/autoscribe: build/autoscribe
	cp build/autoscribe $@

/etc/autoscribe/prompts:
	mkdir -p $@

