package log

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/stablecog/loki-client-go/loki"
)

var logger *log.Logger
var lokiWriter *LokiWriter

func CloseLoki() {
    if lokiWriter != nil && lokiWriter.Client != nil {
        lokiWriter.Client.Stop()
    }
}

func getLogger() *log.Logger {
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

    if logger == nil {
        styles := log.DefaultStyles()
        styles.Levels[log.FatalLevel] = lipgloss.NewStyle().SetString("☠️🟥☠️")
        styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().SetString("🟥")
        styles.Levels[log.WarnLevel] = lipgloss.NewStyle().SetString("🟨")
        styles.Levels[log.InfoLevel] = lipgloss.NewStyle().SetString("")
        logger = log.New(lokiWriter)
        logger.SetStyles(styles)
        /* logger.SetReportTimestamp(true) */
    }
    return logger
}

func Info(msg interface{}, keyvals ...interface{}) {
    getLogger().Info(msg, keyvals...)
}

func Infof(format string, args ...any) {
    getLogger().Infof(format, args...)
}

func Error(msg interface{}, keyvals ...interface{}) {
    getLogger().Error(msg, keyvals...)
}

func Errorf(format string, args ...any) {
    getLogger().Errorf(format, args...)
}

func Warn(msg interface{}, keyvals ...interface{}) {
    getLogger().Warn(msg, keyvals...)
}

func Warnf(format string, args ...any) {
    getLogger().Warnf(format, args...)
}

func Fatal(msg interface{}, keyvals ...interface{}) {
    getLogger().Fatal(msg, keyvals...)
}

func Fatalf(format string, args ...any) {
    getLogger().Fatalf(format, args...)
}