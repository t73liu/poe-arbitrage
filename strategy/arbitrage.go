package strategy

import (
	"errors"
	"fmt"
	"poe-arbitrage/api"
)

// Represents an adjacency list for a directed graph (potentially cyclic)
type TradingPaths struct {
	TradingPairTrades map[TradingPair][]api.TradeDetail
	ItemTradingPairs  map[string][]TradingPair
}

func NewTradingPaths() *TradingPaths {
	return &TradingPaths{
		TradingPairTrades: make(map[TradingPair][]api.TradeDetail),
		ItemTradingPairs:  make(map[string][]TradingPair),
	}
}

type TradingPair struct {
	InitialItem string
	TargetItem  string
}

func (tp *TradingPaths) Set(initialItem, targetItem string, tradeDetails *[]api.TradeDetail) error {
	if initialItem == targetItem {
		return errors.New("invalid trading path: initialItem cannot equal targetItem")
	}

	tradingPair := TradingPair{
		InitialItem: initialItem,
		TargetItem:  targetItem,
	}

	if _, ok := tp.TradingPairTrades[tradingPair]; !ok {
		tradingPairs, _ := tp.ItemTradingPairs[initialItem]
		tp.ItemTradingPairs[initialItem] = append(tradingPairs, tradingPair)
	}
	tp.TradingPairTrades[tradingPair] = *tradeDetails

	return nil
}

func (tp *TradingPaths) Get(initialItem, targetItem string) *[]api.TradeDetail {
	tradingPair := TradingPair{
		InitialItem: initialItem,
		TargetItem:  targetItem,
	}
	bulkTrades, ok := tp.TradingPairTrades[tradingPair]
	if ok {
		return &bulkTrades
	} else {
		return nil
	}
}

// TODO analyze profitable trading paths given initial capital constraints
func (tp *TradingPaths) Analyze(capital map[string]int) error {
	fmt.Printf("%+v\n", tp)
	return nil
}
