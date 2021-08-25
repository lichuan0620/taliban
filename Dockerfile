FROM golang:1.16-stretch as builder

COPY . /go/src/github.com/lichuan0620/taliban
WORKDIR /go/src/github.com/lichuan0620/taliban
RUN go build -o taliban .

FROM debian:stretch-slim

RUN mkdir /taliban && chown -R nobody:nogroup /taliban
COPY --from=builder /go/src/github.com/lichuan0620/taliban/taliban /usr/local/bin/taliban
COPY examples /taliban/examples
COPY LICENSE /taliban/LICENSE

USER       nobody
WORKDIR    /taliban
ENTRYPOINT ["taliban"]
