package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestLogType struct {
	Level   string `json:"level"`
	Message string `json:"message"`

	Time   string      `json:"time,omitempty"`
	Caller string      `json:"caller,omitempty"`
	Stack  interface{} `json:"stack,omitempty"`
	Error  string      `json:"error,omitempty"`

	WithField1 interface{} `json:"with_field_1,omitempty"`
}

func testTextPlainEnable(buf *bytes.Buffer) Option {
	return func(l *Logger) {
		zl := l.l.Output(zerolog.ConsoleWriter{Out: buf}).With().Logger()
		l.l = &zl
	}
}

func testJsonEnable(buf *bytes.Buffer) Option {
	return func(l *Logger) {
		zl := l.l.Output(buf).With().Logger()
		l.l = &zl
	}
}

func getTestData() (context.Context, *bytes.Buffer) {
	var buf = new(bytes.Buffer)
	logger := New(DebugLevel, testJsonEnable(buf))
	return ToContext(context.Background(), logger), buf
}

func getTestDataTextPlain() (context.Context, *bytes.Buffer) {
	var buf = new(bytes.Buffer)
	logger := New(DebugLevel, testTextPlainEnable(buf))
	return ToContext(context.Background(), logger), buf
}

func newTestLogType(data []byte) *TestLogType {
	log := new(TestLogType)
	_ = json.Unmarshal(data, log)
	return log
}

func Test_JsonLog(t *testing.T) {
	t.Parallel()

	t.Run("debug", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Debug(ctx, "debugMsg")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(DebugLevel).String(), log.Level)
		assert.Equal(t, "debugMsg", log.Message)
	})

	t.Run("debug with KV", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Debug(ctx, "debugMsg", "with_field_1", "test")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(DebugLevel).String(), log.Level)
		assert.Equal(t, "debugMsg", log.Message)
		assert.Equal(t, "test", log.WithField1)
	})

	t.Run("info", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Info(ctx, "infoMsg")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(InfoLevel).String(), log.Level)
		assert.Equal(t, "infoMsg", log.Message)
	})

	t.Run("info with KV", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Info(ctx, "infoMsg", "with_field_1", "test")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(InfoLevel).String(), log.Level)
		assert.Equal(t, "infoMsg", log.Message)
		assert.Equal(t, "test", log.WithField1)
	})

	t.Run("warn", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Warn(ctx, "warnMsg")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(WarnLevel).String(), log.Level)
		assert.Equal(t, "warnMsg", log.Message)
	})

	t.Run("warn with KV", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		Warn(ctx, "warnMsg", "with_field_1", "test")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(WarnLevel).String(), log.Level)
		assert.Equal(t, "warnMsg", log.Message)
		assert.Equal(t, "test", log.WithField1)
	})

	t.Run("error", func(t *testing.T) {
		// Arrange
		ctx, buf := getTestData()

		// Act
		Error(ctx, errors.New("test error"), "errorMsg")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(ErrorLevel).String(), log.Level)
		assert.Equal(t, "errorMsg", log.Message)
		assert.Equal(t, "test error", log.Error)
		assert.Equal(t, `<nil>`, fmt.Sprintf("%v", log.Stack))
	})

	t.Run("error with stack", func(t *testing.T) {
		// Arrange
		ctx, buf := getTestData()

		// Act
		SetStacktraceEnabled(true)
		Error(ctx, errors.New("test error"), "errorMsg")
		SetStacktraceEnabled(false)

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(ErrorLevel).String(), log.Level)
		assert.Equal(t, "errorMsg", log.Message)
		assert.Equal(t, "test error", log.Error)
		assert.Contains(t, fmt.Sprintf("%v", log.Stack), `Test_JsonLog`)
		assert.Contains(t, fmt.Sprintf("%v", log.Stack), `source:logger_test.go`)
	})

	t.Run("panic", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		var recovered bool
		func(ctxWithLogger context.Context) {
			defer func() {
				if r := recover(); r != any(nil) {
					recovered = true
				}
			}()
			Panic(ctxWithLogger, "panicMsg")
		}(ctx)

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(PanicLevel).String(), log.Level)
		assert.Equal(t, "panicMsg", log.Message)
		assert.True(t, recovered)
	})

	t.Run("panic with KV", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestData()

		// Act
		var recovered bool
		func(ctxWithLogger context.Context) {
			defer func() {
				if r := recover(); r != any(nil) {
					recovered = true
				}
			}()
			Panic(ctxWithLogger, "panicMsg", "with_field_1", "test")
		}(ctx)

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, zerolog.Level(PanicLevel).String(), log.Level)
		assert.Equal(t, "panicMsg", log.Message)
		assert.Equal(t, "test", log.WithField1)
		assert.True(t, recovered)
	})
}

func Test_TextPlainLog(t *testing.T) {
	t.Parallel()

	t.Run("info", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestDataTextPlain()

		// Act
		Info(ctx, "infoMsg")

		// Assert
		assert.Contains(t, buf.String(), "INF")
		assert.Contains(t, buf.String(), "infoMsg")
	})

	t.Run("debug", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestDataTextPlain()

		// Act
		Debug(ctx, "debugMsg")

		// Assert
		assert.Contains(t, buf.String(), "DBG")
		assert.Contains(t, buf.String(), "debugMsg")
	})

	t.Run("warn", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestDataTextPlain()

		// Act
		Warn(ctx, "warnMsg")

		// Assert
		assert.Contains(t, buf.String(), "WRN")
		assert.Contains(t, buf.String(), "warnMsg")
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestDataTextPlain()

		// Act
		Error(ctx, errors.New("test error"), "errorMsg")

		// Assert
		assert.Contains(t, buf.String(), "ERR")
		assert.Contains(t, buf.String(), "errorMsg")
		assert.Contains(t, buf.String(), `test error`)
	})

	t.Run("panic", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ctx, buf := getTestDataTextPlain()

		// Act
		var recovered bool
		func(ctxWithLogger context.Context) {
			defer func() {
				if r := recover(); r != any(nil) {
					recovered = true
				}
			}()
			Panic(ctxWithLogger, "panicMsg")
		}(ctx)

		// Assert
		assert.Contains(t, buf.String(), "PNC")
		assert.Contains(t, buf.String(), "panicMsg")
		assert.True(t, recovered)
	})
}

func Test_With(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, buf := getTestData()
	logger := FromContext(ctx).With("a", 1, "b", "1", "c", []int{1, 2, 3})
	ctx = ToContext(ctx, logger)

	// Act
	Info(ctx, "infoMsg 1")
	Info(ctx, "infoMsg 2")

	// Assert
	logs := strings.Split(buf.String(), "\n")
	require.Len(t, logs, 3)
	assert.Contains(t, logs[0], `level":"info","a":1,"b":"1","c":[1,2,3]`)
	assert.Contains(t, logs[0], `"message":"infoMsg 1"`)
	assert.Contains(t, logs[1], `level":"info","a":1,"b":"1","c":[1,2,3]`)
	assert.Contains(t, logs[1], `"message":"infoMsg 2"`)
}

func Test_Level(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, buf := getTestData()
	logger := FromContext(ctx).Level(InfoLevel)
	ctx = ToContext(ctx, logger)

	// Act
	Debug(ctx, "hidden")
	Info(ctx, "visible")

	// Assert
	log := buf.String()
	assert.Contains(t, log, "visible")
	assert.NotContains(t, log, "hidden")
}

func Test_Caller(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, buf := getTestData()

	t.Run("enabled", func(t *testing.T) {
		// Act
		SetCallerEnabled(true)
		Info(ctx, "infoMsg")
		SetCallerEnabled(false)

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Contains(t, log.Caller, `/logger_test.go:`)
		buf.Reset()
	})

	t.Run("disabled", func(t *testing.T) {
		// Act
		Info(ctx, "infoMsg")

		// Assert
		log := newTestLogType(buf.Bytes())
		assert.Equal(t, log.Caller, ``)
	})
}
