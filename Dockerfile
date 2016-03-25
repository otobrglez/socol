FROM golang:1.6-onbuild

ADD . /go/src/app

RUN go get github.com/tools/godep && \
  godep restore && \
  godep go build && \
  godep go install && \
  go get ./...

EXPOSE 5000

ENTRYPOINT ["/go/bin/socol", "-s"]
