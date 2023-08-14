FROM node AS node
WORKDIR /app
ADD ui /app/ui
RUN ls
RUN cd /app/ui/; npm install; npm run build

FROM golang AS golang
WORKDIR /app
ADD . /app/
COPY --from=node /app/ui/build/ /app/ui/build/
RUN CGO_ENABLED=0 go build -o /ybFeed *.go

FROM scratch
COPY --from=golang /ybFeed /ybFeed

EXPOSE 8080

ENTRYPOINT [ "/ybFeed" ]