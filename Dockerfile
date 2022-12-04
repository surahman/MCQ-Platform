# syntax=docker/dockerfile:1

## README
# Please review the configuration files in the /configs directory and adjust configurations
# accordingly.


## Build
FROM golang:1.18-alpine AS build

RUN apk update && apk --no-cache --update add build-base

WORKDIR /build

COPY . ./
RUN make dep
RUN make build ARCH=linux

## Deploy
FROM alpine:latest

WORKDIR /

COPY --from=build /build/MCQPlatform-linux /MCQPlatform-linux
COPY --from=build /build/configs /etc/MCQ_Platform.conf/


# Port list:
# Please set these ports according to the configurations in the YAML files in /configs directory.
# 1) HTTP REST
# 2) HTTP GraphQL
# 3) Cassandra
# 4) Cassandra
# 5) Redis-node-0
# 6) Redis-node-1
# 7) Redis-node-2
# 8) Redis-node-3
# 9) Redis-node-4
# 10) Redis-node-5
EXPOSE 44243 44255 7000 9042 6379 6380 6381 6382 6383 6384

# Launch application.
ENTRYPOINT ["./MCQPlatform-linux"]
