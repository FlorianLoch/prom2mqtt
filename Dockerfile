FROM golang:1.21.4 AS builder

WORKDIR /src/
# Copy dependency management related files first and download required modules before copying changed code into the
# container. That way we can cache the downloading as long as the dependency configuration does not change too.
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o prom2mqtt .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=builder /src/prom2mqtt ./
CMD ["./prom2mqtt"]
