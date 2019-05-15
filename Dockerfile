<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> bd72195... Importing basic server logic for https and jwt
FROM golang:1.12.5 as builder

WORKDIR /go/app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM scratch
COPY --from=builder /go/app/main /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

VOLUME ["/certs"]

<<<<<<< HEAD
ENTRYPOINT [ "/app/main" ]
=======
ARG GO_VERSION=1.12.4
FROM golang:${GO_VERSION}-alpine3.9 as golang

ENV SRC_DIR=/go/src
WORKDIR ${SRC_DIR}
COPY ./ ./

# RUN go get all
# RUN go mod verify
# RUN go mod tidy
RUN go build -o gopro .
ENTRYPOINT [ "./gopro", "-env=local" ]
EXPOSE 8085
EXPOSE 80
>>>>>>> e38123e... go mod verify
=======
ENTRYPOINT [ "/app/main" ]
>>>>>>> bd72195... Importing basic server logic for https and jwt
