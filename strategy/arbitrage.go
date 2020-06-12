package strategy

import (
	"errors"
	"fmt"
	"poe-arbitrage/api"
	"poe-arbitrage/utils"
	"strconv"
	"strings"
)

// Represents an adjacency list for a directed graph (potentially cyclic)
type TradingPaths struct {
	tradingPairTrades     map[TradingPair][]api.TradeDetail
	itemTradingPairs      map[string][]TradingPair
	capital               map[string]int
	noCapitalRequirements bool
}

func NewTradingPaths(capital map[string]int) *TradingPaths {
	return &TradingPaths{
		tradingPairTrades:     make(map[TradingPair][]api.TradeDetail),
		itemTradingPairs:      make(map[string][]TradingPair),
		capital:               capital,
		noCapitalRequirements: len(capital) == 0,
	}
}

type TradingPair struct {
	InitialItem string
	TargetItem  string
}

type ValidTrade struct {
	listing api.TradeDetail
	whisper string
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

func (tp *TradingPaths) Analyze() error {
	initialItems := make([]string, 0, len(tp.itemTradingPairs))

	// Filter out invalid starting trades based on capital
	for item := range tp.itemTradingPairs {
		if tp.noCapitalRequirements {
			initialItems = append(initialItems, item)
		} else if _, ok := tp.capital[item]; ok {
			initialItems = append(initialItems, item)
		}
	}

	for _, item := range initialItems {
		fmt.Println("Trades starting from:", item)
		initialPairs := tp.itemTradingPairs[item]
		tradingCycles := make([][]TradingPair, 0, len(initialPairs))

		// Check for direct trade paths (e.g. exa => chaos => exa)
		for _, initialTradePair := range initialPairs {
			reverseTradePair := TradingPair{
				InitialItem: initialTradePair.TargetItem,
				TargetItem:  initialTradePair.InitialItem,
			}
			if _, ok := tp.tradingPairTrades[reverseTradePair]; ok {
				tradingCycles = append(tradingCycles, []TradingPair{
					initialTradePair,
					reverseTradePair,
				})
			}
		}

		// TODO: Check simple cycle trade paths (e.g. exa => gcp => chaos => exa)
		// stack := make([]TradingPair, len(initialPairs))
		// DFS with backtrack and visited set

		// Print profitable trade paths
		for _, tradingCycle := range tradingCycles {
			tp.printProfitableTradePath(tradingCycle)
		}
	}
	return nil
}

func (tp *TradingPaths) printProfitableTradePath(tradingCycle []TradingPair) {
	validTrades := make([]ValidTrade, 0, 5)
	initialPair := tradingCycle[0]
	initialItem := initialPair.InitialItem
	initialAmount := uint(tp.capital[initialItem])

	if tp.noCapitalRequirements {
		initialTrade := tp.tradingPairTrades[initialPair][0]
		initialAmount = initialTrade.Stock
	}
	currentAmount := initialAmount
	hypotheticalPnL := 100.0
	for _, pair := range tradingCycle {
		trades := tp.tradingPairTrades[pair]
		noValidTrades := true
		for _, trade := range trades {
			if trade.PriceAmount <= currentAmount {
				noValidTrades = false
				maxPrice, maxItem := calcMaxTransaction(
					trade.PriceAmount,
					trade.ItemAmount,
					trade.Stock,
					currentAmount,
				)
				currentAmount = maxItem
				hypotheticalPnL = trade.Ratio * hypotheticalPnL
				validTrades = append(
					validTrades,
					ValidTrade{
						listing: trade,
						whisper: formatWhisper(trade.Whisper, maxPrice, maxItem),
					},
				)
				break
			}
		}
		// If a single trading pair fails then stop evaluating the rest of the cycle
		if noValidTrades {
			break
		}
	}
	// At least 5% gain
	if hypotheticalPnL > 101 && len(validTrades) == len(tradingCycle) {
		for _, validTrade := range validTrades {
			fmt.Println(validTrade.whisper)
			printTradeDetail(validTrade.listing)
		}
		fmt.Printf("\nGains: %.3f%%\n\n", hypotheticalPnL)
	}
	fmt.Println()
}

// Assumes that capital satisfies initial price and calculates the max item amount
// that can be purchased
func calcMaxTransaction(priceAmount, itemAmount, stockSize, capital uint) (maxPrice, maxItem uint) {
	gcd := utils.CalcGCD(priceAmount, itemAmount)
	minPrice := priceAmount / gcd
	minItem := itemAmount / gcd

	maxNumberOfSales := stockSize / minItem
	maxNumberOfPurchases := capital / minPrice

	maxNumberOfTransactions := utils.CalcMin(maxNumberOfPurchases, maxNumberOfSales)

	maxPrice = maxNumberOfTransactions * minPrice
	maxItem = maxNumberOfTransactions * minItem
	return maxPrice, maxItem
}

func formatWhisper(whisper string, priceAmount, itemAmount uint) string {
	whisper = strings.Replace(whisper, "{0}", strconv.Itoa(int(priceAmount)), 1)
	whisper = strings.Replace(whisper, "{1}", strconv.Itoa(int(itemAmount)), 1)
	return whisper
}

func printTradeDetail(tradeDetail api.TradeDetail) {
	fmt.Println("Pay:", tradeDetail.PriceAmount, tradeDetail.PriceUnit)
	fmt.Println("Receive:", tradeDetail.ItemAmount, tradeDetail.ItemUnit)
	fmt.Println("Stock:", tradeDetail.Stock)
	fmt.Println("Account:", tradeDetail.Account)
}
