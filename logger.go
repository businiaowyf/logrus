package logrus

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	// The logs are `io.Copy`'d to this in a mutex. It's common to set this to a
	// file, or leave it default which is `os.Stderr`. You can also set this to
	// something more adventorous, such as logging to Kafka.
	Out io.Writer
	// Hooks for the logger instance. These allow firing events based on logging
	// levels and log entries. For example, to send errors to an error tracking
	// service, log to StatsD or dump the core on fatal errors.
	Hooks LevelHooks
	// All log entries pass through the formatter before logged to Out. The
	// included formatters are `TextFormatter` and `JSONFormatter` for which
	// TextFormatter is the default. In development (when a TTY is attached) it
	// logs with colors, but to a file it wouldn't. You can easily implement your
	// own that implements the `Formatter` interface, see the `README` or included
	// formatters for examples.
	Formatter Formatter
	// The logging level the logger should log at. This is typically (and defaults
	// to) `logrus.Info`, which allows Info(), Warn(), Error() and Fatal() to be
	// logged.
	Level Level
	// Used to sync writing to the log. Locking is enabled by Default
	mu MutexWrap
	// Reusable empty entry
	entryPool sync.Pool
}

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}

// Creates a new logger. Configuration should be set by changing `Formatter`,
// `Out` and `Hooks` directly on the default logger instance. You can also just
// instantiate your own:
//
//    var log = &Logger{
//      Out: os.Stderr,
//      Formatter: new(JSONFormatter),
//      Hooks: make(LevelHooks),
//      Level: logrus.DebugLevel,
//    }
//
// It's recommended to make this a global instance called `log`.
func New() *Logger {
	return &Logger{
		Out:       os.Stderr,
		Formatter: new(TextFormatter),
		Hooks:     make(LevelHooks),
		Level:     InfoLevel,
	}
}

func (logger *Logger) newEntry() *Entry {
	entry, ok := logger.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(logger)
}

func (logger *Logger) releaseEntry(entry *Entry) {
	entry.Data = map[string]interface{}{}
	logger.entryPool.Put(entry)
}

// Adds a field to the log entry, note that it doesn't log until you call
// Debug, Print, Info, Warn, Error, Fatal or Panic. It only creates a log entry.
// If you want multiple fields, use `WithFields`.
func (logger *Logger) WithField(key string, value interface{}) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithField(key, value)
}

// Adds a struct of fields to the log entry. All it does is call `WithField` for
// each `Field`.
func (logger *Logger) WithFields(fields Fields) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithFields(fields)
}

// Add an error as single field to the log entry.  All it does is call
// `WithError` for the given `error`.
func (logger *Logger) WithError(err error) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithError(err)
}

// Overrides the time of the log entry.
func (logger *Logger) WithTime(t time.Time) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithTime(t)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.level() >= DebugLevel {
		entry := logger.newEntry()
		entry.debugf(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.level() >= InfoLevel {
		entry := logger.newEntry()
		entry.infof(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.printf(format, args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warnf(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warnf(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	if logger.level() >= ErrorLevel {
		entry := logger.newEntry()
		entry.errorf(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	if logger.level() >= FatalLevel {
		entry := logger.newEntry()
		entry.fatalf(format, args...)
		logger.releaseEntry(entry)
	}
	Exit(1)
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	if logger.level() >= PanicLevel {
		entry := logger.newEntry()
		entry.panicf(format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Debug(args ...interface{}) {
	if logger.level() >= DebugLevel {
		entry := logger.newEntry()
		entry.debug(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Info(args ...interface{}) {
	if logger.level() >= InfoLevel {
		entry := logger.newEntry()
		entry.info(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Print(args ...interface{}) {
	entry := logger.newEntry()
	entry.info(args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warn(args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warn(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Warning(args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warn(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Error(args ...interface{}) {
	if logger.level() >= ErrorLevel {
		entry := logger.newEntry()
		entry.error(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatal(args ...interface{}) {
	if logger.level() >= FatalLevel {
		entry := logger.newEntry()
		entry.fatal(args...)
		logger.releaseEntry(entry)
	}
	Exit(1)
}

func (logger *Logger) Panic(args ...interface{}) {
	if logger.level() >= PanicLevel {
		entry := logger.newEntry()
		entry.panic(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Debugln(args ...interface{}) {
	if logger.level() >= DebugLevel {
		entry := logger.newEntry()
		entry.debugln(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Infoln(args ...interface{}) {
	if logger.level() >= InfoLevel {
		entry := logger.newEntry()
		entry.infoln(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Println(args ...interface{}) {
	entry := logger.newEntry()
	entry.println(args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnln(args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warnln(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Warningln(args ...interface{}) {
	if logger.level() >= WarnLevel {
		entry := logger.newEntry()
		entry.warnln(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Errorln(args ...interface{}) {
	if logger.level() >= ErrorLevel {
		entry := logger.newEntry()
		entry.errorln(args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatalln(args ...interface{}) {
	if logger.level() >= FatalLevel {
		entry := logger.newEntry()
		entry.fatalln(args...)
		logger.releaseEntry(entry)
	}
	Exit(1)
}

func (logger *Logger) Panicln(args ...interface{}) {
	if logger.level() >= PanicLevel {
		entry := logger.newEntry()
		entry.panicln(args...)
		logger.releaseEntry(entry)
	}
}

//When file is opened with appending mode, it's safe to
//write concurrently to a file (within 4k message on Linux).
//In these cases user can choose to disable the lock.
func (logger *Logger) SetNoLock() {
	logger.mu.Disable()
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

func (logger *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

func (logger *Logger) SetOutput(out io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Out = out
}

func (logger *Logger) AddHook(hook Hook) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Hooks.Add(hook)
}
