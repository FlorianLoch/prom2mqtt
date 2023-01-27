# The Raspberry will require an ARM image, so the platform building the image should be ARM too. If this
# is not possible, additioniol measures will probably be required - like building with `docker buildx`
# Using an Alpine based image for the build is not possible, we need the full toolchain because of cgo.
FROM golang:1.19.5-buster AS builder

WORKDIR /src/
# Copy dependency management related files first and download required modules before copying changed code into the
# container. That way we can cache the downloading as long as the dependency configuration does not change too.
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o prom2mqtt .


FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=builder /src/prom2mqtt ./
CMD ["./prom2mqtt"]
