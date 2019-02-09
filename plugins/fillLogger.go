package plugins

import (
	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/logger"
)

// FillLogger is a FillHandler that logs fills
type FillLogger struct {
	l logger.Logger
}

var _ api.FillHandler = &FillLogger{}

// MakeFillLogger is a factory method
func MakeFillLogger(l logger.Logger) api.FillHandler {
	return &FillLogger{
		l,
	}
}

// HandleFill impl.
func (f *FillLogger) HandleFill(trade model.Trade) error {
	f.l.Infof("received fill: %s\n", trade)
	return nil
}
