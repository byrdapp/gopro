FROM golang:1.14 as builder

WORKDIR /go/app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build /go/app/cmd/byrd-pro-api/main.go

FROM scratch
COPY --from=builder /go/app/main/ /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "/app/main" ]
