# syntax=docker/dockerfile:1

# https://hub.docker.com/r/hadolint/hadolint/tags
FROM hadolint/hadolint:v2.12.0-alpine AS hadolint
# https://hub.docker.com/r/golangci/golangci-lint/tags
FROM golangci/golangci-lint:v2.1.6-alpine AS golangci
# https://hub.docker.com/r/transformia/task/tags
FROM transformia/task:3.43.2 AS task
# https://hub.docker.com/_/golang
FROM golang:1.25.0-alpine3.21
COPY --from=hadolint /bin/hadolint /bin/hadolint
COPY --from=golangci /usr/bin/golangci-lint /bin/golangci-lint-v2
COPY --from=task /usr/local/bin/task /bin/task
# bash is required for clive
# hadolint ignore=DL3018
RUN apk add --no-cache bash git gcc linux-headers musl-dev docker-cli openssh-client htop && \ 
  ln -s /bin/golangci-lint-v2 /bin/golangci-lint
