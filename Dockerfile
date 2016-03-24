FROM golang:1.6-onbuild

WORKDIR /go/src/github.com/otobrglez/socol

ADD . /go/src/github.com/otobrglez/socol

RUN go get ./... && \
  go get github.com/tools/godep && \
  godep restore && \
  godep go build && \
  godep go install

EXPOSE 5000

CMD ["socol -s -p 5000"]
