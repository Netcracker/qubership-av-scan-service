FROM golang:1.24.9-alpine3.21 AS builder

WORKDIR /go/src/av-scan-service
COPY go.mod go.sum main.go ./
COPY pkg ./pkg/
RUN go build -o av-scan-service . 

FROM alpine:3.21

RUN apk update && apk upgrade

COPY --from=builder /go/src/av-scan-service/av-scan-service /av-scan-service

ENV USER_UID=2001 \
    USER_NAME=av-scan-service \
    GROUP_NAME=av-scan-service
RUN addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}