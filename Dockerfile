FROM golang:1.24.0-alpine3.21 AS builder

WORKDIR /go/src/github.com/aifoundry-org/storage-manager
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/storage-manager

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/storage-manager /storage-manager

EXPOSE 8050

ENTRYPOINT [ "/storage-manager" ]
