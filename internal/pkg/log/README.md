# 패키지 구조

### logger.go
- Logger 인터페이스 정의
- context 처리 시 개별 메소드에서 받는 것 권장
	- 수명 문제 -> context는 요청 단위로 짧게 유지하는 것 권장
	- 동시성 문제 -> 고루틴 등에서 여러 컨텍스트가 생겨날 수도 있음

### zap.go 등의 구현체
- 로깅 라이브러리별로 Logger 인터페이스를 구현한 파일

### context
- log 패키지에서 context 처리하기 위한 파일
- context에서 string을 key로 사용하는 것은 지양해야한다고 함
- 때문에 context를 int key로 처리하기 위한 기능이 포함되어 있음

=> 추후 공통패키지로 빼는 것 고려

### (Deprecated) factory.go 
- 로깅 라이브러리를 선택하여 인스턴스 및 실제 구현체를 반환하는 파일
- 인터페이스보다 실제 구현체를 반환하는 편이 좋다고 들었음
    - but ZapLogging도 구현체
    - 불필요한 LoggerWrapper 필요없다고 판단되어 삭제

```go
package log

import "context"

type LoggerWrapper struct {
	logger Logger
}

func NewLoggerWrapper(loggerType string, ctx context.Context) *LoggerWrapper {
	var logger Logger

	switch loggerType {
	case "zap":
		logger = NewZapLogger(ctx)
	default:
		logger = NewZapLogger(ctx)
	}

	return &LoggerWrapper{logger: logger}
}

func (lw *LoggerWrapper) Debug(msg string, fields map[string]interface{}) {

	lw.logger.Debug(msg, fields)
}

func (lw *LoggerWrapper) Info(msg string, fields map[string]interface{}) {
	lw.logger.Info(msg, fields)
}

func (lw *LoggerWrapper) Warn(msg string, fields map[string]interface{}) {
	lw.logger.Warn(msg, fields)
}

func (lw *LoggerWrapper) Error(msg string, fields map[string]interface{}) {
	lw.logger.Error(msg, fields)
}

func (lw *LoggerWrapper) Fatal(msg string, fields map[string]interface{}) {
	lw.logger.Fatal(msg, fields)
}

func (lw *LoggerWrapper) Clean() error {
	return lw.logger.Clean()
}

```