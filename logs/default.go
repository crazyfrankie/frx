package logs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

var logger FullLogger = &defaultLogger{
	level:  LevelInfo,
	stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
}

// SetOutput sets the output of default logs. By default, it is stderr.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelTrace.
// Note that this method is not concurrent-safe.
func SetLevel(lv Level) {
	logger.SetLevel(lv)
}

// DefaultLogger return the default logs for kitex.
func DefaultLogger() FullLogger {
	return logger
}

// SetLogger sets the default logs.
// Note that this method is not concurrent-safe and must not be called
// after the use of DefaultLogger and global functions in this package.
func SetLogger(v FullLogger) {
	logger = v
}

// Fatal calls the default logs's Fatal method and then os.Exit(1).
func Fatal(v ...any) {
	logger.Fatal(v...)
}

// Error calls the default logs's Error method.
func Error(v ...any) {
	logger.Error(v...)
}

// Warn calls the default logs's Warn method.
func Warn(v ...any) {
	logger.Warn(v...)
}

// Notice calls the default logs's Notice method.
func Notice(v ...any) {
	logger.Notice(v...)
}

// Info calls the default logs's Info method.
func Info(v ...any) {
	logger.Info(v...)
}

// Debug calls the default logs's Debug method.
func Debug(v ...any) {
	logger.Debug(v...)
}

// Trace calls the default logs's Trace method.
func Trace(v ...any) {
	logger.Trace(v...)
}

// Fatalf calls the default logs's Fatalf method and then os.Exit(1).
func Fatalf(format string, v ...any) {
	logger.Fatalf(format, v...)
}

// Errorf calls the default logs's Errorf method.
func Errorf(format string, v ...any) {
	logger.Errorf(format, v...)
}

// Warnf calls the default logs's Warnf method.
func Warnf(format string, v ...any) {
	logger.Warnf(format, v...)
}

// Noticef calls the default logs's Noticef method.
func Noticef(format string, v ...any) {
	logger.Noticef(format, v...)
}

// Infof calls the default logs's Infof method.
func Infof(format string, v ...any) {
	logger.Infof(format, v...)
}

// Debugf calls the default logs's Debugf method.
func Debugf(format string, v ...any) {
	logger.Debugf(format, v...)
}

// Tracef calls the default logs's Tracef method.
func Tracef(format string, v ...any) {
	logger.Tracef(format, v...)
}

// CtxFatalf calls the default logs's CtxFatalf method and then os.Exit(1).
func CtxFatalf(ctx context.Context, format string, v ...any) {
	logger.CtxFatalf(ctx, format, v...)
}

// CtxErrorf calls the default logs's CtxErrorf method.
func CtxErrorf(ctx context.Context, format string, v ...any) {
	logger.CtxErrorf(ctx, format, v...)
}

// CtxWarnf calls the default logs's CtxWarnf method.
func CtxWarnf(ctx context.Context, format string, v ...any) {
	logger.CtxWarnf(ctx, format, v...)
}

// CtxNoticef calls the default logs's CtxNoticef method.
func CtxNoticef(ctx context.Context, format string, v ...any) {
	logger.CtxNoticef(ctx, format, v...)
}

// CtxInfof calls the default logs's CtxInfof method.
func CtxInfof(ctx context.Context, format string, v ...any) {
	logger.CtxInfof(ctx, format, v...)
}

// CtxDebugf calls the default logs's CtxDebugf method.
func CtxDebugf(ctx context.Context, format string, v ...any) {
	logger.CtxDebugf(ctx, format, v...)
}

// CtxTracef calls the default logs's CtxTracef method.
func CtxTracef(ctx context.Context, format string, v ...any) {
	logger.CtxTracef(ctx, format, v...)
}

type defaultLogger struct {
	stdlog *log.Logger
	level  Level
}

func (ll *defaultLogger) SetOutput(w io.Writer) {
	ll.stdlog.SetOutput(w)
}

func (ll *defaultLogger) SetLevel(lv Level) {
	ll.level = lv
}

func (ll *defaultLogger) logf(lv Level, format *string, v ...any) {
	if ll.level > lv {
		return
	}
	msg := lv.toString()
	if format != nil {
		msg += fmt.Sprintf(*format, v...)
	} else {
		msg += fmt.Sprint(v...)
	}
	ll.stdlog.Output(4, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (ll *defaultLogger) logfCtx(ctx context.Context, lv Level, format *string, v ...any) {
	if ll.level > lv {
		return
	}
	msg := lv.toString()
	logID := ctx.Value("log-id")
	if logID != nil {
		msg += fmt.Sprintf("[log-id: %v] ", logID)
	}
	if format != nil {
		msg += fmt.Sprintf(*format, v...)
	} else {
		msg += fmt.Sprint(v...)
	}
	ll.stdlog.Output(4, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (ll *defaultLogger) Fatal(v ...any) {
	ll.logf(LevelFatal, nil, v...)
}

func (ll *defaultLogger) Error(v ...any) {
	ll.logf(LevelError, nil, v...)
}

func (ll *defaultLogger) Warn(v ...any) {
	ll.logf(LevelWarn, nil, v...)
}

func (ll *defaultLogger) Notice(v ...any) {
	ll.logf(LevelNotice, nil, v...)
}

func (ll *defaultLogger) Info(v ...any) {
	ll.logf(LevelInfo, nil, v...)
}

func (ll *defaultLogger) Debug(v ...any) {
	ll.logf(LevelDebug, nil, v...)
}

func (ll *defaultLogger) Trace(v ...any) {
	ll.logf(LevelTrace, nil, v...)
}

func (ll *defaultLogger) Fatalf(format string, v ...any) {
	ll.logf(LevelFatal, &format, v...)
}

func (ll *defaultLogger) Errorf(format string, v ...any) {
	ll.logf(LevelError, &format, v...)
}

func (ll *defaultLogger) Warnf(format string, v ...any) {
	ll.logf(LevelWarn, &format, v...)
}

func (ll *defaultLogger) Noticef(format string, v ...any) {
	ll.logf(LevelNotice, &format, v...)
}

func (ll *defaultLogger) Infof(format string, v ...any) {
	ll.logf(LevelInfo, &format, v...)
}

func (ll *defaultLogger) Debugf(format string, v ...any) {
	ll.logf(LevelDebug, &format, v...)
}

func (ll *defaultLogger) Tracef(format string, v ...any) {
	ll.logf(LevelTrace, &format, v...)
}

func (ll *defaultLogger) CtxFatalf(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelFatal, &format, v...)
}

func (ll *defaultLogger) CtxErrorf(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelError, &format, v...)
}

func (ll *defaultLogger) CtxWarnf(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelWarn, &format, v...)
}

func (ll *defaultLogger) CtxNoticef(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelNotice, &format, v...)
}

func (ll *defaultLogger) CtxInfof(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelInfo, &format, v...)
}

func (ll *defaultLogger) CtxDebugf(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelDebug, &format, v...)
}

func (ll *defaultLogger) CtxTracef(ctx context.Context, format string, v ...any) {
	ll.logfCtx(ctx, LevelTrace, &format, v...)
}
