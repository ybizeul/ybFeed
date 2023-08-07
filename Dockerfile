FROM node AS node
WORKDIR /app
ADD ui ui
RUN ls
RUN cd /app/ui/; npm install; npm run build

FROM golang AS golang
WORKDIR /app
ADD . /app/
COPY --from=node /app/ui/build/ /app/ui/
RUN go build -o ybFeed *.go

FROM scratch
COPY --from=golang /app/ybFeed /ybFeed
ADD ybFeed /ybFeed

ENTRYPOINT [ "/ybFeed" ]