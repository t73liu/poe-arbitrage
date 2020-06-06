package strategy

import (
	"errors"
	"fmt"
	"poe-arbitrage/api"
)

// Represents an adjacency list for a directed graph (potentially cyclic)
type TradingPaths struct {
	tradingPairTrades map[TradingPair][]api.TradeDetail
	itemTradingPairs  map[string][]TradingPair
}

func NewTradingPaths() *TradingPaths {
	return &TradingPaths{
		tradingPairTrades: make(map[TradingPair][]api.TradeDetail),
		itemTradingPairs:  make(map[string][]TradingPair),
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

	if _, ok := tp.tradingPairTrades[tradingPair]; !ok {
		tradingPairs, _ := tp.itemTradingPairs[initialItem]
		tp.itemTradingPairs[initialItem] = append(tradingPairs, tradingPair)
	}
	tp.tradingPairTrades[tradingPair] = *tradeDetails

	return nil
}

func (tp *TradingPaths) Get(initialItem, targetItem string) *[]api.TradeDetail {
	tradingPair := TradingPair{
		InitialItem: initialItem,
		TargetItem:  targetItem,
	}
	bulkTrades, ok := tp.tradingPairTrades[tradingPair]
	if ok {
		return &bulkTrades
	} else {
		return nil
	}
}

// TODO analyze profitable trading paths given initial capital constraints
func (tp *TradingPaths) Analyze(initialCapital map[string]int) error {
	for tradingPair, trades := range tp.tradingPairTrades {
		fmt.Printf("%+v\n", tradingPair)
		printTradeDetail(trades[0])
		fmt.Println()
	}
	// Filter out invalid starting trades based on capital
	// Find valid trading pairs via ItemTradingPairs
	// initial trade can be dependent on one player
	// subsequent trades should require multiple players as backup
	// Check if "cycle" (typically a cycle requires 3 nodes) is lucrative
	// If the first trade for the pair does not work, subsequent trades will not either
	// If the initial capital cannot satisfy first trade, look at subsequent trades and alert
	return nil
}

func printTradeDetail(tradeDetail api.TradeDetail) {
	fmt.Println("Pay:", tradeDetail.PriceAmount, tradeDetail.PriceUnit)
	fmt.Println("Receive:", tradeDetail.ItemAmount, tradeDetail.ItemUnit)
	fmt.Println("Stock:", tradeDetail.Stock)
	fmt.Println("Account:", tradeDetail.Account)
}
