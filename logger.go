package logger

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Level int8

const (
	DebugLevel = Level(zerolog.DebugLevel)
	InfoLevel  = Level(zerolog.InfoLevel)
	WarnLevel  = Level(zerolog.WarnLevel)
	ErrorLevel = Level(zerolog.ErrorLevel)
	FatalLevel = Level(zerolog.FatalLevel)
	PanicLevel = Level(zerolog.PanicLevel)

	minAllowedLevel = DebugLevel
	maxAllowedLevel = PanicLevel
)

var (
	stacktraceEnabled atomic.Bool
	callerEnabled     atomic.Bool
)

type Logger struct {
	l *zerolog.Logger
}

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.DefaultContextLogger = New(ErrorLevel).l
}

func New(level Level, opts ...Option) *Logger {
	zl := zerolog.New(os.Stdout)
	zl = zl.Level(zerolog.Level(level))
	zl = zl.With().Timestamp().Logger()

	l := &Logger{l: &zl}
	for _, opt := range opts {
		opt(l)
	}

	return l
}

func SetGlobalLevel(level Level) {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

func SetStacktraceEnabled(enabled bool) {
	stacktraceEnabled.Store(enabled)
}

func SetCallerEnabled(enabled bool) {
	callerEnabled.Store(enabled)
}

func (l *Logger) Level(level Level) *Logger {
	zl := l.l.Level(zerolog.Level(level))
	return &Logger{l: &zl}
}

func (l *Logger) With(kvs ...interface{}) *Logger {
	zl := l.l.With().Fields(kvs).Logger()
	return &Logger{l: &zl}
}

func (l *Logger) Debug(msg string, kvs ...interface{}) {
	event := l.l.Debug()
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func (l *Logger) Info(msg string, kvs ...interface{}) {
	event := l.l.Info()
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func (l *Logger) Warn(msg string, kvs ...interface{}) {
	event := l.l.Warn()
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func (l *Logger) Error(err error, msg string, kvs ...interface{}) {
	event := l.l.Error()

	if stacktraceEnabled.Load() {
		event = event.Stack()
	}

	event = event.Err(err)
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, kvs ...interface{}) {
	event := l.l.Fatal()
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func (l *Logger) Panic(msg string, kvs ...interface{}) {
	event := l.l.Panic()
	event = withFieldsAndCaller(event, kvs...)
	event.Msg(msg)
}

func Debug(ctx context.Context, msg string, kvs ...interface{}) {
	FromContext(ctx).Debug(msg, kvs...)
}

func Info(ctx context.Context, msg string, kvs ...interface{}) {
	FromContext(ctx).Info(msg, kvs...)
}

func Warn(ctx context.Context, msg string, kvs ...interface{}) {
	FromContext(ctx).Warn(msg, kvs...)
}

func Error(ctx context.Context, err error, msg string, kvs ...interface{}) {
	FromContext(ctx).Error(err, msg, kvs...)
}

func Fatal(ctx context.Context, msg string, kvs ...interface{}) {
	FromContext(ctx).Fatal(msg, kvs...)
}

func Panic(ctx context.Context, msg string, kvs ...interface{}) {
	FromContext(ctx).Panic(msg, kvs...)
}

func ParseLevel(levelStr string) (Level, error) {
	zeroLevel, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return ErrorLevel, fmt.Errorf("cannot parse log level \"%s\" %w", levelStr, err)
	}

	level := Level(zeroLevel)

	if level < minAllowedLevel {
		fmt.Printf("log level \"%s\" less than min allowed\n", levelStr)
		return minAllowedLevel, nil
	}
	if level > maxAllowedLevel {
		fmt.Printf("log level \"%s\" greater than max allowed\n", levelStr)
		return maxAllowedLevel, nil
	}

	return level, nil
}

func withFieldsAndCaller(event *zerolog.Event, kvs ...interface{}) *zerolog.Event {
	if len(kvs) > 0 {
		event = event.Fields(kvs)
	}

	if callerEnabled.Load() {
		event = event.Caller(3)
	}

	return event
}
