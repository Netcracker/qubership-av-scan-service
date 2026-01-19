FROM golang:1.24.11-alpine3.23.2.01 AS builder

WORKDIR /go/src/av-scan-service
COPY go.mod go.sum main.go ./
COPY pkg ./pkg/
RUN go build -o av-scan-service . 

FROM alpine:3.23.2

RUN apk update && apk upgrade

COPY --from=builder /go/src/av-scan-service/av-scan-service /av-scan-service

ENV USER_UID=2001 \
    USER_NAME=av-scan-service \
    GROUP_NAME=av-scan-service
RUN addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}