# 디렉토리 구조

## internal 최상위
- 내부적으로만 사용되나 한 곳에서만 호출되는 패키지의 경우 internal 최상위에 위치
- ex) config 패키지
    - cmd/entrypoint/main.go에서만 호출되는 구조

## internal/app
- 내가 작성한 어플리케이션 패키지 및 코드를 모아놓은 패키지

## internal/pkg
- 애플리케이션 코드에서 공유되는 코드들을 모아놓은 패키지
- logging, HTTP server 등