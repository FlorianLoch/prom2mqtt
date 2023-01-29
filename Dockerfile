FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /
COPY ./bin/prom2mqtt-arm32v7 ./
CMD ["./prom2mqtt"]
