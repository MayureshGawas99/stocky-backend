package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()


func Init() {
	Log.SetOutput(os.Stdout)

	service := os.Getenv("LOG_SERVICE")
	if service == "" {
		service = filepath.Base(os.Args[0])
	}

	Log.SetReportCaller(true)
	Log.SetFormatter(&ServiceFormatter{Service: service})

	// Log level
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
}

type ServiceFormatter struct{
	Service string
}

// ANSI color codes
const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorGreen  = "\x1b[32m"
	colorCyan   = "\x1b[36m"
	colorBlue   = "\x1b[34m"
)

func (f *ServiceFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ts := entry.Time.UTC().Format(time.RFC3339)
	level := strings.ToUpper(entry.Level.String())

	forceColor := strings.ToLower(os.Getenv("LOG_FORCE_COLOR"))
	colorEnabled := forceColor == "1" || forceColor == "true" || isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	if colorEnabled {
		level = fmt.Sprintf("%s%s%s", colorForLevel(entry.Level), level, colorReset)
	}

	file := "unknown"
	if entry.HasCaller() && entry.Caller != nil {
		file = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
	}

	line := fmt.Sprintf("[%s] [%s] [%s] %s\n", level, ts, file, entry.Message)
	return []byte(line), nil
}

func colorForLevel(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return colorCyan
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	case logrus.InfoLevel:
		return colorGreen
	default:
		return colorBlue
	}
}

