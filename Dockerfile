FROM golang:1.26-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o manticore ./cmd/manticore

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /build/manticore /usr/local/bin/manticore

ENTRYPOINT ["manticore"]
