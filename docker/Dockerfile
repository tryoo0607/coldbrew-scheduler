############## BUILD ##############

# syntax=docker/dockerfile:1

FROM golang:1.25.0 AS builder
WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64


# deps 캐시
# COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

# 앱 소스
COPY . .

RUN go build -trimpath -ldflags="-s -w" -o /coldbrew-scheduler-build ./cmd/entrypoint

############## RUN ##############
FROM scratch
# FROM busybox:1.36

WORKDIR /
COPY --from=builder /coldbrew-scheduler-build /coldbrew-scheduler

ENTRYPOINT ["/coldbrew-scheduler"]
