package logger

import (
	"time"

	"go.uber.org/zap"
)

type Config struct {
	// FileName is the file to write logs to.  Backup log files will be retained
	// in the same directory, if fileName is blank it will not write the log to a file.
	FileName string
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes
	MaxSize int
	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.
	MaxAge int
	// WithCaller is the flag to enable caller in log
	WithCaller bool
	// CallerSkip is the number of stack frames to skip to find the caller
	CallerSkip int
	// BufferSize is the number of buffer size to cache the log in memory
	BufferSize int
	// FlushBufferInterval specifies how often the writer should flush data if
	// there have been no writes.
	FlushBufferInterval time.Duration
	// Encoding json || console
	Encoding string
	// SkipStdOutput is true. Logger will not write to standard output
	SkipStdOutput bool
	// Level default is debug
	Level string
	// if set ZapLogger. Logger will use this instance to implementation
	ZapLogger *zap.Logger
}
