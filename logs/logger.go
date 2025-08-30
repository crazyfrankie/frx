package logs

import (
	"context"
	"fmt"
	"io"
)

// FormatLogger is a logs interface that output logs with a format.
type FormatLogger interface {
	Tracef(format string, v ...any)
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Noticef(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

// Logger is a logs interface that provides logging function with levels.
type Logger interface {
	Trace(v ...any)
	Debug(v ...any)
	Info(v ...any)
	Notice(v ...any)
	Warn(v ...any)
	Error(v ...any)
	Fatal(v ...any)
}

// CtxLogger is a logs interface that accepts a context argument and output
// logs with a format.
type CtxLogger interface {
	CtxTracef(ctx context.Context, format string, v ...any)
	CtxDebugf(ctx context.Context, format string, v ...any)
	CtxInfof(ctx context.Context, format string, v ...any)
	CtxNoticef(ctx context.Context, format string, v ...any)
	CtxWarnf(ctx context.Context, format string, v ...any)
	CtxErrorf(ctx context.Context, format string, v ...any)
	CtxFatalf(ctx context.Context, format string, v ...any)
}

// Control provides methods to config a logs.
type Control interface {
	SetLevel(Level)
	SetOutput(io.Writer)
}

// FullLogger is the combination of Logger, FormatLogger, CtxLogger and Control.
type FullLogger interface {
	Logger
	FormatLogger
	CtxLogger
	Control
}

// Level defines the priority of a log message.
// When a logs is configured with a level, any log message with a lower
// log level (smaller by integer comparison) will not be output.
type Level int

// The levels of logs.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarn
	LevelError
	LevelFatal
)

var strs = []string{
	"[Trace] ",
	"[Debug] ",
	"[Info] ",
	"[Notice] ",
	"[Warn] ",
	"[Error] ",
	"[Fatal] ",
}

func (lv Level) toString() string {
	if lv >= LevelTrace && lv <= LevelFatal {
		return strs[lv]
	}
	return fmt.Sprintf("[?%d] ", lv)
}
