FROM wehkamp/alpine:3.4

LABEL container.name="wehkamp/prometheus-node-exporter"

ENV  GOPATH /go
ENV APPPATH $GOPATH/src/github.com/prometheus/node_exporter

WORKDIR $APPPATH

ADD . $APPPATH/

RUN apk add --update -t go && \
	go build -o /bin/node-exporter

EXPOSE      9100
ENTRYPOINT ["/bin/node-exporter", "-collector.filesystem.ignored-mount-points", ".*/(sys|proc|dev|etc|docker)($|/)", "-collector.ntp.server", "3.amazon.pool.ntp.org"]
