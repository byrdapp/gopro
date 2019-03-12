FROM golang:1.12.0-alpine3.9 as golang
LABEL version="1.0"
ENV SRC_DIR=/go/src/github.com/byblix/gopro

ADD . ${SRC_DIR}
WORKDIR ${SRC_DIR}

RUN go get -d ./...
RUN go build -o gopro .
ENTRYPOINT [ "./gopro", "-env=local" ]
EXPOSE 8085
