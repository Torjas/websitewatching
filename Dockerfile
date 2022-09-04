#build stage
FROM golang:1.18-buster AS  builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /docker-watcher

#final stage
FROM ubuntu:bionic

WORKDIR /
ARG apt_sources="http://archive.ubuntu.com"

RUN sed -i "s|http://archive.ubuntu.com|$apt_sources|g" /etc/apt/sources.list && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
    # chromium dependencies
    libnss3 \
    libxss1 \
    libasound2 \
    libxtst6 \
    libgtk-3-0 \
    libgbm1 \
    ca-certificates \
    # fonts
    fonts-liberation fonts-noto-color-emoji fonts-noto-cjk \
    # timezone
    tzdata \
    # processs reaper
    dumb-init \
    # headful mode support, for example: $ xvfb-run chromium-browser --remote-debugging-port=9222
    xvfb \
    # cleanup
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /docker-watcher /docker-watcher



ENTRYPOINT ["/docker-watcher"]
