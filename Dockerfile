FROM wehkamp/alpine:3.4

LABEL container.name="wehkamp/prometheus-node-exporter:latest"

ENV  GOPATH /go
ENV APPPATH $GOPATH/src/github.com/prometheus/node_exporter

WORKDIR $APPPATH


ADD . $APPPATH/

RUN apk add --update -t build-deps go libc-dev gcc libgcc && \
	go build -o /bin/node-exporter && \
	apk del --purge build-deps && rm -rf $GOPATH

EXPOSE      9100
ENTRYPOINT ["/bin/node-exporter", "-collector.filesystem.ignored-mount-points", ".*/(sys|proc|dev|etc|docker)($|/)"]
