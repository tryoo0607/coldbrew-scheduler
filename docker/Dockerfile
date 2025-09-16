############## BUILD ##############
# syntax=docker/dockerfile:1

FROM golang:1.25.0 AS builder
WORKDIR /app

# deps 캐시
COPY go.mod ./
RUN go mod download

# 앱 소스
COPY . .

# Makefile을 이용해 빌드
RUN make build

############## RUN ##############
FROM scratch
# FROM busybox:1.36

WORKDIR /
COPY --from=builder /app/bin/coldbrew-scheduler /coldbrew-scheduler

ENTRYPOINT ["/coldbrew-scheduler"]
