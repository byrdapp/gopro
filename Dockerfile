FROM golang:1.12.0-alpine3.9 as golang
LABEL version="1.0"
ENV SRC_DIR=/go/src/github.com/byblix/gopro
ENV GO111MODULE=on

ADD . ${SRC_DIR}
WORKDIR ${SRC_DIR}

RUN ls -a
RUN apk update && apk upgrade && apk add --no-cache bash git openssh
RUN go mod verify
RUN go mod download
RUN go build -o gopro .
ENTRYPOINT [ "./gopro", "-env=local" ]
EXPOSE 8085
