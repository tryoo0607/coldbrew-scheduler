package log

import (
	"context"

	"go.uber.org/zap"
)

type ZapLogger struct {
 logger *zap.Logger
 ctx    context.Context
}

func NewZapLogger(loggerType string, ctx context.Context) *ZapLogger {
 logger, _ := zap.NewProduction()

 return &ZapLogger{logger: logger, ctx: ctx}
}

func (l *ZapLogger) Debug(msg string, fields map[string]interface{}) {
 l.addContextCommonFields(fields)

 l.logger.Debug("", zap.Any("args", fields))
}

func (l *ZapLogger) Info(msg string, fields map[string]interface{}) {
 l.addContextCommonFields(fields)

 l.logger.Info("", zap.Any("args", fields))
}

func (l *ZapLogger) Warn(msg string, fields map[string]interface{}) {
 l.addContextCommonFields(fields)

 l.logger.Warn("", zap.Any("args", fields))
}

func (l *ZapLogger) Error(msg string, fields map[string]interface{}) {
 l.addContextCommonFields(fields)

 l.logger.Error("", zap.Any("args", fields))
}

func (l *ZapLogger) Fatal(msg string, fields map[string]interface{}) {
 l.addContextCommonFields(fields)

 l.logger.Fatal("", zap.Any("args", fields))
}

func (l *ZapLogger) addContextCommonFields(fields map[string]interface{}) {
 if l.ctx != nil {
  for k, v := range l.ctx.Value("commonFields").(map[string]interface{}) {
   if _, ok := fields[k]; !ok {
    fields[k] = v
   }
  }
 }
}