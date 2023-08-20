all: ui go push

ui:
	cd ui; \
	npm install; \
	npm run build

go:
	go build -o ybFeed cmd/ybfeed/*.go

ui-run: ui run

run:
	go run cmd/ybfeed/*.go

push:
	ko build -B -t `git describe --tags`

clean:
	rm -f ybFeed
	rm -rf ui/node_modules

.PHONY: ui
