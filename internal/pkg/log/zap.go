package log

import (
	"context"

	"go.uber.org/zap"
)

type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger() *ZapLogger {
	logger, _ := zap.NewProduction()

	return &ZapLogger{l: logger}
}

// 공통 필드를 바인딩한 "자식 로거" 반환
func (z *ZapLogger) With(f Fields) Logger {
	if len(f) == 0 {
		return z
	}
	return &ZapLogger{l: z.l.With(toFields(f)...)}
}

func (z *ZapLogger) Debug(ctx context.Context, msg string, f Fields) { z.log(ctx, z.l.Debug, msg, f) }
func (z *ZapLogger) Info(ctx context.Context, msg string, f Fields)  { z.log(ctx, z.l.Info, msg, f) }
func (z *ZapLogger) Warn(ctx context.Context, msg string, f Fields)  { z.log(ctx, z.l.Warn, msg, f) }
func (z *ZapLogger) Error(ctx context.Context, msg string, f Fields) { z.log(ctx, z.l.Error, msg, f) }
func (z *ZapLogger) Fatal(ctx context.Context, msg string, f Fields) { z.log(ctx, z.l.Fatal, msg, f) }
func (z *ZapLogger) Sync() error                                     { return z.l.Sync() }

func (z *ZapLogger) log(ctx context.Context, fn func(string, ...zap.Field), msg string, f Fields) {
	cf := fieldsFromCtx(ctx)
	merged := merge(cf, f) // nil 안전, 얕은 복사-병합
	if len(merged) == 0 {
		fn(msg)
		return
	}
	fn(msg, toFields(merged)...)
}

func toFields(m Fields) []zap.Field {
	if len(m) == 0 {
		return nil
	}
	fs := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fs = append(fs, zap.Any(k, v))
	}
	return fs
}

func merge(a, b Fields) Fields {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	if len(a) == 0 {
		return copyFields(b)
	}
	if len(b) == 0 {
		return copyFields(a)
	}
	out := make(Fields, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	} // b가 a를 덮어씀
	return out
}

func copyFields(m Fields) Fields {
	if len(m) == 0 {
		return nil
	}
	out := make(Fields, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
