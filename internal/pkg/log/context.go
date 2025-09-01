package log

import "context"

// typed key (문자열 키 금지)
type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyUserID
)

func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, keyRequestID, id)
}
func WithUserID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, keyUserID, id)
}

// 로거가 매 호출마다 꺼내서 병합
func fieldsFromCtx(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}
	out := Fields{}
	if v := ctx.Value(keyRequestID); v != nil {
		if s, ok := v.(string); ok && s != "" {
			out["request_id"] = s
		}
	}
	if v := ctx.Value(keyUserID); v != nil {
		if s, ok := v.(string); ok && s != "" {
			out["user_id"] = s
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
