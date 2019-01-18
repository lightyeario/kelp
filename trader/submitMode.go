package trader

import (
	"fmt"

	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/plugins"
	"github.com/stellar/go/build"
)

// SubmitMode is the type of mode to be used when submitting orders to the trader bot
type SubmitMode uint8

// constants for the SubmitMode
const (
	SubmitModeMakerOnly SubmitMode = iota
	SubmitModeTakerOnly
	SubmitModeBoth
)

// ParseSubmitMode converts a string to the SubmitMode constant
func ParseSubmitMode(submitMode string) SubmitMode {
	if submitMode == "maker_only" {
		return SubmitModeMakerOnly
	}

	if submitMode == "taker_only" {
		return SubmitModeTakerOnly
	}

	return SubmitModeBoth
}

// submitFilter allows you to filter out operations before submitting to the network
type submitFilter interface {
	apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error)
}

// makeSubmitFilter makes a submit filter based on the passed in submitMode
func makeSubmitFilter(submitMode SubmitMode, sdex *plugins.SDEX, tradingPair *model.TradingPair) submitFilter {
	if submitMode == SubmitModeMakerOnly || submitMode == SubmitModeTakerOnly {
		return &sdexFilter{
			tradingPair: tradingPair,
			sdex:        sdex,
			submitMode:  submitMode,
		}
	}
	return nil
}

type sdexFilter struct {
	tradingPair *model.TradingPair
	sdex        *plugins.SDEX
	submitMode  SubmitMode
}

var _ submitFilter = &sdexFilter{}

func (f *sdexFilter) apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error) {
	ob := &model.OrderBook{}
	// we only want the top bid and ask values so use a maxCount of 1
	// ob, e := f.sdex.GetOrderBook(f.tradingPair, 1)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not fetch SDEX orderbook: %s", e)
	// }
	var e error

	// TODO find intersection of orderbook and ops
	/*
		1. get top bid and top ask in OB
		2. for each op remove or keep op if it is before/after top bid/ask depending on the mode we're in
	*/

	if f.submitMode == SubmitModeMakerOnly {
		ops, e = filterMakerMode(ops, ob)
		if e != nil {
			return nil, fmt.Errorf("could not apply maker mode filter: %s", e)
		}
	} else {
		ops, e = filterTakerMode(ops, ob)
		if e != nil {
			return nil, fmt.Errorf("could not apply taker mode filter: %s", e)
		}
	}
	return ops, nil
}

func filterMakerMode(ops []build.TransactionMutator, ob *model.OrderBook) ([]build.TransactionMutator, error) {
	return nil, nil
}

func filterTakerMode(ops []build.TransactionMutator, ob *model.OrderBook) ([]build.TransactionMutator, error) {
	return nil, nil
}
