package log

import "context"

type Fields = map[string]interface{}

type Logger interface {
	With(Fields) Logger
	Debug(ctx context.Context, msg string, fields Fields)
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, fields Fields)
	Fatal(ctx context.Context, msg string, fields Fields)
	Sync() error
}
