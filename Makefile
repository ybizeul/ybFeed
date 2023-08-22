VERSION = $(shell git describe --tags)

all: ui go push

ui:
	cd web/ui/; \
	npm install; \
	npm run build

go:
	GOFLAGS="-ldflags=-X=main.version=$(VERSION)" \
	go build -o ybFeed cmd/ybfeed/*.go

ui-run: ui run

run:
	GOFLAGS="-ldflags=-X=main.version=$(VERSION)" \
	go run cmd/ybfeed/*.go

push:
	GOFLAGS="-ldflags=-X=main.version=$(VERSION)" ko build -B -t $(VERSION)

clean:
	rm -f ybFeed
	rm -rf web/ui/node_modules

.PHONY: ui
