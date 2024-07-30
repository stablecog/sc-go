package log

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/stablecog/loki-client-go/loki"
)

var infoLogger *log.Logger
var warnLogger *log.Logger
var errorLogger *log.Logger
var fatalLogger *log.Logger
var lokiWriter *LokiWriter

func CloseLoki() {
	if lokiWriter != nil && lokiWriter.Client != nil {
		lokiWriter.Client.Stop()
	}
}

func getLogger(level log.Level) *log.Logger {
	if lokiWriter == nil {
		lokiApplicationLabel := os.Getenv("LOKI_APPLICATION_LABEL")
		if lokiApplicationLabel == "" {
			lokiApplicationLabel = "sc-go"
		}
		lokiPushUrl := os.Getenv("LOKI_PUSH_URL")
		if lokiPushUrl != "" {
			config, _ := loki.NewDefaultConfig(lokiPushUrl)
			lokiClient, err := loki.New(config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create loki client: %v", err)
				panic(err)
			}
			lokiWriter = &LokiWriter{
				Stderr:               os.Stderr,
				Client:               lokiClient,
				LokiApplicationLabel: lokiApplicationLabel,
			}
		} else {
			lokiWriter = &LokiWriter{
				Stderr: os.Stderr,
			}
		}
	}

	if level == log.FatalLevel {
		if fatalLogger == nil {
			fatalLogger = log.New(lokiWriter)
			fatalLogger.SetPrefix("☠️🟥☠️")
			/* fatalLogger.SetReportTimestamp(true) */
		}
		return fatalLogger
	}
	if level == log.ErrorLevel {
		if errorLogger == nil {
			errorLogger = log.New(lokiWriter)
			errorLogger.SetPrefix("🟥")
			/* errorLogger.SetReportTimestamp(true) */
		}
		return errorLogger
	}
	if level == log.WarnLevel {
		if warnLogger == nil {
			warnLogger = log.New(lokiWriter)
			warnLogger.SetPrefix("🟨")
			/* warnLogger.SetReportTimestamp(true) */
		}
		return warnLogger
	}
	if infoLogger == nil {
		infoLogger = log.New(lokiWriter)
		infoLogger.SetPrefix("🟦")
		/* infoLogger.SetReportTimestamp(true) */
	}
	return infoLogger
}

func Info(msg interface{}, keyvals ...interface{}) {
	getLogger(log.InfoLevel).Info(msg, keyvals...)
}

func Infof(format string, args ...any) {
	getLogger(log.InfoLevel).Info(fmt.Sprintf(format, args...))
}

func Error(msg interface{}, keyvals ...interface{}) {
	getLogger(log.ErrorLevel).Error(msg, keyvals...)
}

func Errorf(format string, args ...any) {
	getLogger(log.ErrorLevel).Error(fmt.Sprintf(format, args...))
}

func Warn(msg interface{}, keyvals ...interface{}) {
	getLogger(log.WarnLevel).Warn(msg, keyvals...)
}

func Warnf(format string, args ...any) {
	getLogger(log.WarnLevel).Warn(fmt.Sprintf(format, args...))
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	getLogger(log.FatalLevel).Fatal(msg, keyvals...)
}

func Fatalf(format string, args ...any) {
	getLogger(log.FatalLevel).Fatal(fmt.Sprintf(format, args...))
}
