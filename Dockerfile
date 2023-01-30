FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY ./bin/prom2mqtt-arm32v7 ./prom2mqtt
CMD ["./prom2mqtt"]
