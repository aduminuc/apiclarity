FROM golang:1.16.6-alpine AS builder

RUN apk add --update --no-cache gcc g++

ARG WORKING_DIR

WORKDIR /plugins
COPY plugins/api ./api

WORKDIR /plugins/gateway/kong
COPY plugins/gateway/kong/go.* ./
RUN go mod download

COPY plugins/gateway/kong .
# Build the plugin.
RUN apk add make && make all

FROM busybox
#
COPY --from=builder ["/plugins/gateway/kong/bin/kong-plugin", "/usr/local/bin/kong-plugin"]
