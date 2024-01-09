VERSION = $(shell git describe --tags)

all: ui go

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

push: ui
	GOFLAGS="-ldflags=-X=main.version=$(VERSION)" ko build -B -t latest ./cmd/ybfeed/

clean:
	rm -f ybFeed
	rm -rf web/ui/node_modules
	rm -rf web/ui/dist

.PHONY: ui
