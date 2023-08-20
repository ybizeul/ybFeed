FROM node AS node
WORKDIR /app/web
ADD web/ui /app/web/ui
RUN cd /app/web/ui/; npm install; npm run build

FROM golang AS golang
WORKDIR /app
ADD . /app/
COPY --from=node /app/web/ui/build/ /app/web/ui/build/
RUN CGO_ENABLED=0 go build -o /ybFeed cmd/ybfeed/*.go

FROM scratch
COPY --from=golang /ybFeed /ybFeed

EXPOSE 8080

ENTRYPOINT [ "/ybFeed" ]

LABEL "org.opencontainers.image.authors"="yann@tynsoe.org"