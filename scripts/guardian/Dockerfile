ARG GO_VERSION=1.26.6
FROM golang:${GO_VERSION}-bookworm
RUN apt-get update \
    && apt-get install --yes --no-install-recommends python3 git ca-certificates \
    && rm -rf /var/lib/apt/lists/*
