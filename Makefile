all: ui go

ui:
	cd ui; \
	npm install; \
	npm run build

go:
	go build -o ybFeed *.go

run:
	go run *.go

clean:
	rm -f ybFeed
	rm -rf ui/node_modules

.PHONY: ui
