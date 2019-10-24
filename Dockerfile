FROM golang:1.13 as builder

WORKDIR /go/app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build /go/app/cmd/gopro/gopro.go

FROM scratch
COPY --from=builder /go/app/gopro/ /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# VOLUME ["/certs"]

EXPOSE 8080

ENTRYPOINT [ "/app/gopro" ]