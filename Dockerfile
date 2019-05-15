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

EXPOSE 443
EXPOSE 80

ENTRYPOINT [ "/app/main" ]