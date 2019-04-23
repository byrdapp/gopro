ARG GO_VERSION=1.12.4
FROM golang:${GO_VERSION}-alpine3.9 as golang

ENV SRC_DIR=/go/src
WORKDIR ${SRC_DIR}
COPY ./ ./

RUN go mod download
RUN go mod verify
RUN go mod tidy
RUN go build -o gopro .
ENTRYPOINT [ "./gopro", "-env=local" ]
EXPOSE 8085
EXPOSE 80
