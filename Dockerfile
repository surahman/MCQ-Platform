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


# Please set the LHS to match the port you would like exposed externally and the RHS to match the
# internal HTTP ports as set in the configuration YAML files.

# HTTP REST endpoint.
EXPOSE 44243:44243
# HTTP GraphQL endpoint.
EXPOSE 44255:44255

# Launch application.
ENTRYPOINT ["./MCQPlatform-linux"]
