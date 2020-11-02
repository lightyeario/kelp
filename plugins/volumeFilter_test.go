package plugins

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/utils"

	"github.com/stellar/go/txnbuild"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

func makeTestVolumeFilterConfig(baseCapInBase, baseCapInQuote float64, additionalMarketIDs, optionalAccountIDs []string, mode volumeFilterMode) *VolumeFilterConfig {
	var baseCapInBasePtr *float64
	if baseCapInBase >= 0 {
		baseCapInBasePtr = pointy.Float64(baseCapInBase)
	}

	var baseCapInQuotePtr *float64
	if baseCapInQuote >= 0 {
		baseCapInQuotePtr = pointy.Float64(baseCapInQuote)
	}

	return &VolumeFilterConfig{
		SellBaseAssetCapInBaseUnits:  baseCapInBasePtr,
		SellBaseAssetCapInQuoteUnits: baseCapInQuotePtr,
		mode:                         mode,
		additionalMarketIDs:          additionalMarketIDs,
		optionalAccountIDs:           optionalAccountIDs,
	}
}

func makeWantVolumeFilter(config *VolumeFilterConfig, firstMarketID string, marketIDs []string, optionalAccountIDs []string, action string) *volumeFilter {
	queryMarketIDs := utils.Dedupe(append([]string{firstMarketID}, marketIDs...))
	query, e := queries.MakeDailyVolumeByDateForMarketIdsAction(&sql.DB{}, queryMarketIDs, action, optionalAccountIDs)
	if e != nil {
		panic(e)
	}

	return &volumeFilter{
		name:                   "volumeFilter",
		baseAsset:              utils.NativeAsset,
		quoteAsset:             utils.NativeAsset,
		config:                 config,
		dailyVolumeByDateQuery: query,
	}
}

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.MakeSdexMappedAssetDisplayFn(map[model.Asset]hProtocol.Asset{model.Asset("XLM"): utils.NativeAsset})
	configValue := ""
	exchangeName := ""
	tradingPair := &model.TradingPair{Base: "XLM", Quote: "XLM"}
	modes := []volumeFilterMode{volumeFilterModeExact, volumeFilterModeIgnore}
	firstMarketID := MakeMarketID(exchangeName, "native", "native")

	testCases := []struct {
		name       string
		marketIDs  []string
		accountIDs []string
		wantFilter *volumeFilter
	}{
		// TODO DS Confirm the empty config fails once validation is added to the constructor
		{
			name:       "1 market id",
			marketIDs:  []string{"marketID"},
			accountIDs: []string{},
		},
		{
			name:       "2 market ids",
			marketIDs:  []string{"marketID1", "marketID2"},
			accountIDs: []string{},
		},
		{
			name:       "2 dupe market ids, 1 distinct",
			marketIDs:  []string{"marketID1", "marketID1", "marketID2"},
			accountIDs: []string{},
		},
		{
			name:       "1 account id",
			marketIDs:  []string{},
			accountIDs: []string{"accountID"},
		},
		{
			name:       "2 account ids",
			marketIDs:  []string{},
			accountIDs: []string{"accountID1", "accountID2"},
		},
		{
			name:       "account and market ids",
			marketIDs:  []string{"marketID"},
			accountIDs: []string{"accountID"},
		},
	}

	for _, k := range testCases {
		// this lets us test both types of modes when varying the market and account ids
		for _, m := range modes {
			// this lets us test both constraints within the config
			baseCapInBaseConfig := makeTestVolumeFilterConfig(1.0, -1.0, k.marketIDs, k.accountIDs, m)
			baseCapInQuoteConfig := makeTestVolumeFilterConfig(-1.0, 1.0, k.marketIDs, k.accountIDs, m)

			for _, config := range []*VolumeFilterConfig{baseCapInBaseConfig, baseCapInQuoteConfig} {
				// configType is used to represent the type of config when printing test name
				var configType string
				if config.SellBaseAssetCapInBaseUnits != nil {
					configType = "base"
				} else {
					configType = "quote"
				}

				// TODO DS Vary filter action between buy and sell, once buy logic is implemented.
				wantFilter := makeWantVolumeFilter(config, firstMarketID, k.marketIDs, k.accountIDs, "sell")
				t.Run(fmt.Sprintf("%s/%s/%s", k.name, configType, m), func(t *testing.T) {
					actual, e := makeFilterVolume(
						configValue,
						exchangeName,
						tradingPair,
						testAssetDisplayFn,
						utils.NativeAsset,
						utils.NativeAsset,
						&sql.DB{},
						config,
					)

					if !assert.Nil(t, e) {
						return
					}

					assert.Equal(t, wantFilter, actual)
				})
			}
		}
	}
}

func makeManageSellOffer(price, amount string) *txnbuild.ManageSellOffer {
	if amount == "" {
		return nil
	}

	return &txnbuild.ManageSellOffer{
		Buying:  txnbuild.NativeAsset{},
		Selling: txnbuild.NativeAsset{},
		Price:   price,
		Amount:  amount,
	}
}

func TestVolumeFilterFn(t *testing.T) {
	testCases := []struct {
		name               string
		filter             *volumeFilter
		sellBaseCapInBase  *float64
		sellBaseCapInQuote *float64
		otbBaseCap         float64
		otbQuoteCap        float64
		tbbBaseCap         float64
		tbbQuoteCap        float64
		price              string
		inputAmount        string
		wantAmount         string
		wantTbbBaseCap     float64
		wantTbbQuoteCap    float64
	}{
		{
			name:               "selling, base units sell cap, don't keep selling base",
			sellBaseCapInBase:  pointy.Float64(0.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     0.0,
			wantTbbQuoteCap:    0.0,
		},
		{
			name:               "selling, base units sell cap, keep selling base",
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "1.0000000",
			wantTbbBaseCap:     1.0,
			wantTbbQuoteCap:    2.0,
		},
		{
			name:               "selling, quote units sell cap, don't keep selling quote",
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(0),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     0.0,
			wantTbbQuoteCap:    0.0,
		},
		{
			name:               "selling, quote units sell cap, keep selling quote",
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     0.5,
			wantTbbQuoteCap:    1.0,
		},
		{
			name:               "selling, base and quote units sell cap, keep selling base and quote",
			sellBaseCapInBase:  pointy.Float64(1.),
			sellBaseCapInQuote: pointy.Float64(1.),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     0.5,
			wantTbbQuoteCap:    1.0,
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			marketIDs := []string{}
			accountIDs := []string{}
			mode := volumeFilterModeExact
			dailyOTB := makeTestVolumeFilterConfig(k.otbBaseCap, k.otbQuoteCap, marketIDs, accountIDs, mode)
			dailyTBB := makeTestVolumeFilterConfig(k.tbbBaseCap, k.tbbQuoteCap, marketIDs, accountIDs, mode)
			wantTBB := makeTestVolumeFilterConfig(k.wantTbbBaseCap, k.wantTbbQuoteCap, marketIDs, accountIDs, mode)
			op := makeManageSellOffer(k.price, k.inputAmount)
			wantOp := makeManageSellOffer(k.price, k.wantAmount)

			lp := limitParameters{
				sellBaseAssetCapInBaseUnits:  k.sellBaseCapInBase,
				sellBaseAssetCapInQuoteUnits: k.sellBaseCapInQuote,
				mode:                         volumeFilterModeExact,
			}

			actual, e := volumeFilterFn(dailyOTB, dailyTBB, op, utils.NativeAsset, utils.NativeAsset, lp)

			assert.Nil(t, e)
			assert.Equal(t, wantOp, actual)
			assert.Equal(t, wantTBB, dailyTBB)
		})
	}
}