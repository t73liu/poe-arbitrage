package strategy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/t73liu/poe-arbitrage/api"
	"github.com/t73liu/poe-arbitrage/utils"
)

// TradingPaths represents an adjacency list for a directed graph (potentially cyclic)
type TradingPaths struct {
	tradingPairTrades     map[TradingPair][]api.TradeDetail
	itemTradingPairs      map[string][]TradingPair
	capital               map[string]int
	noCapitalRequirements bool
}

type TradingPair struct {
	InitialItem string
	TargetItem  string
}

type validTrade struct {
	listing api.TradeDetail
	whisper string
}

type tradePathsDFS struct {
	initialItem string
	visited     map[string]bool
	currentPath []TradingPair
	result      [][]TradingPair
}

func NewTradingPaths(capital map[string]int) *TradingPaths {
	return &TradingPaths{
		tradingPairTrades:     make(map[TradingPair][]api.TradeDetail),
		itemTradingPairs:      make(map[string][]TradingPair),
		capital:               capital,
		noCapitalRequirements: len(capital) == 0,
	}
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

	for _, initialItem := range initialItems {
		fmt.Println("Trades starting from:", initialItem)
		dfs := &tradePathsDFS{
			initialItem: initialItem,
			visited:     make(map[string]bool),
			currentPath: make([]TradingPair, 0, len(tp.itemTradingPairs)),
			result:      make([][]TradingPair, 0, len(tp.itemTradingPairs)),
		}
		tp.getTradingPaths(initialItem, dfs)
		for _, tradingPath := range dfs.result {
			tp.printProfitableTradePath(tradingPath)
		}
	}
	return nil
}

// DFS with backtrack and visited set  (e.g. exa => gcp => chaos => exa)
func (tp *TradingPaths) getTradingPaths(item string, dfs *tradePathsDFS) {
	if dfs.visited[item] {
		if item == dfs.initialItem {
			dfs.result = append(dfs.result, dfs.currentPath)
		}
	} else {
		dfs.visited[item] = true
		for _, pair := range tp.itemTradingPairs[item] {
			dfs.currentPath = append(dfs.currentPath, pair)
			tp.getTradingPaths(pair.TargetItem, dfs)
			newPath := make([]TradingPair, len(dfs.currentPath)-1)
			copy(newPath, dfs.currentPath)
			dfs.currentPath = newPath
		}
		dfs.visited[item] = false
	}
}

func (tp *TradingPaths) printProfitableTradePath(tradingPath []TradingPair) {
	validTrades := make([]validTrade, 0, 5)
	initialPair := tradingPath[0]
	initialItem := initialPair.InitialItem
	initialAmount := uint(tp.capital[initialItem])

	if tp.noCapitalRequirements {
		initialTrade := tp.tradingPairTrades[initialPair][0]
		initialAmount = initialTrade.Stock
	}
	currentAmount := initialAmount
	hypotheticalPnL := 100.0

	for _, pair := range tradingPath {
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
				currentAmount = maxItem + uint(tp.capital[pair.TargetItem])
				hypotheticalPnL = trade.Ratio * hypotheticalPnL
				validTrades = append(
					validTrades,
					validTrade{
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

	// At least 1% gain
	if hypotheticalPnL > 101 && len(validTrades) == len(tradingPath) {
		fmt.Printf("%+v\n", tradingPath)
		for _, validTrade := range validTrades {
			fmt.Println(validTrade.whisper)
			printTradeDetail(validTrade.listing)
		}
		fmt.Printf("\nGains: %.3f%% %s\n", hypotheticalPnL-100, initialItem)
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
	whisper = strings.Replace(whisper, "{0}", strconv.Itoa(int(itemAmount)), 1)
	whisper = strings.Replace(whisper, "{1}", strconv.Itoa(int(priceAmount)), 1)
	return whisper
}

func printTradeDetail(tradeDetail api.TradeDetail) {
	fmt.Println("Pay:", tradeDetail.PriceAmount, tradeDetail.PriceUnit)
	fmt.Println("Receive:", tradeDetail.ItemAmount, tradeDetail.ItemUnit)
	fmt.Println("Stock:", tradeDetail.Stock)
	fmt.Printf("Ratio: %.3f\n", tradeDetail.Ratio)
}
