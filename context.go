package logger

import (
	"context"

	"github.com/rs/zerolog"
)

func ToContext(ctx context.Context, logger *Logger) context.Context {
	return logger.l.WithContext(ctx)
}

func FromContext(ctx context.Context) *Logger {
	return &Logger{l: zerolog.Ctx(ctx)}
}

//func loggerFromSpanContext(zl *zerolog.Logger, ctx opentracing.SpanContext) *zerolog.Logger {
//	spanCtx, ok := ctx.(*jaeger.SpanContext)
//	if !ok {
//		return zl
//	}
//
//	logger := zl.With().
//		Str("trace_id", spanCtx.TraceID().String()).
//		Str("span_id", spanCtx.SpanID().String()).
//		Logger()
//
//	return &logger
//}
