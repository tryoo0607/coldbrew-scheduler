APP_NAME := coldbrew-scheduler
DOCKER_USER := tryoo0607
DOCKER_REPO := $(DOCKER_USER)/$(APP_NAME)
BIN_DIR := bin

TAG ?= latest
TAGS ?= $(DOCKER_REPO):$(TAG)

# 빌드 환경 변수
GO111MODULE := on
CGO_ENABLED := 0
GOOS := linux
GOARCH := amd64

## 명령어 인식 처리
.PHONY: build run test docker-build docker-push clean

## Go 빌드
build:
	GO111MODULE=$(GO111MODULE) \
	CGO_ENABLED=$(CGO_ENABLED) \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/entrypoint

## 실행
run: build
	./$(BIN_DIR)/$(APP_NAME)

## 테스트
test:
	go test ./... -v

## Docker 이미지 빌드 (첫 번째 태그만 사용)
docker-build:
	docker build -t $(firstword $(TAGS)) -f ./docker/Dockerfile .

## Docker 이미지 푸시 (모든 태그 처리)
docker-push: docker-build
	@for tag in $(TAGS); do \
		if [ "$$tag" != "$(firstword $(TAGS))" ]; then \
			echo "Tagging $$tag"; \
			docker tag $(firstword $(TAGS)) $$tag; \
		fi; \
		echo "Pushing $$tag"; \
		docker push $$tag; \
	done

## 정리
clean:
	rm -rf $(BIN_DIR)
