#build stage
FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN apk add --no-cache git
RUN go get -d -v ./...
RUN go install -v ./...
RUN ls

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
RUN chgrp -R 0 /app && chmod -R g=u /app
ENTRYPOINT ./app
LABEL Name=mattermost-go-bot Version=0.0.1
EXPOSE 8080
