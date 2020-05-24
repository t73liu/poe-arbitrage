package strategy

import "poe-arbitrage/api"

// Represents an adjacency matrix for a directed graph
type TradingPaths struct {
	Graph map[TradingPair]api.BulkTrades
}

func (tp *TradingPaths) Add(initialItem, targetItem string, bulkTrades api.BulkTrades) {
	tradingPair := TradingPair{
		InitialItem: initialItem,
		TargetItem:  targetItem,
	}
	tp.Graph[tradingPair] = bulkTrades
}

func (tp *TradingPaths) Get(initialItem, targetItem string) *api.BulkTrades {
	tradingPair := TradingPair{
		InitialItem: initialItem,
		TargetItem:  targetItem,
	}
	bulkTrades, ok := tp.Graph[tradingPair]
	if ok {
		return &bulkTrades
	} else {
		return nil
	}
}

type TradingPair struct {
	InitialItem string
	TargetItem  string
}
