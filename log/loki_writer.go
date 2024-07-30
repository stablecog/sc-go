package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stablecog/loki-client-go/loki"
)

type LokiWriter struct {
	Stderr               io.Writer
	Client               *loki.Client
	LokiApplicationLabel string
}

func (lw *LokiWriter) Write(p []byte) (n int, err error) {
	// Write to stderr
	n, err = lw.Stderr.Write(p)
	if err != nil {
		return n, err
	}

	// Write to Loki
	if lw.Client != nil {
		labels := model.LabelSet{
			"application": model.LabelValue(lw.LokiApplicationLabel),
		}
		err = lw.Client.Handle(labels, time.Now(), string(p))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to send log to loki: %v", err)
			return n, err
		}
	}

	return n, nil
}
