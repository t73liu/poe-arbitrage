package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"poe-arbitrage/api"
	"poe-arbitrage/strategy"
	"poe-arbitrage/utils"
	"sort"
	"time"
)

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Check for trading opportunities for bulk items",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("provide at least 2 items")
		}

		if err := validateItems(args, "Invalid arguments: "); err != nil {
			return err
		}

		initialCapital, err := cmd.Flags().GetStringToInt("capital")
		if err != nil {
			fmt.Println("Could not parse --capital argument:", err)
			return err
		}

		capitalItems := make([]string, 0, len(initialCapital))
		for item := range initialCapital {
			capitalItems = append(capitalItems, item)
		}

		if err := validateItems(capitalItems, "Invalid capital: "); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, items []string) error {
		initialCapital, err := cmd.Flags().GetStringToInt("capital")
		if err != nil {
			fmt.Println("Could not parse --capital argument:", err)
			return err
		}

		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println("Failed to parse config:", err)
			return err
		}

		if err := analyzeBulkTrades(items, initialCapital, config); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tradeCmd)

	tradeCmd.Flags().StringToIntP(
		"capital",
		"c",
		make(map[string]int),
		"Specify starting capital (i.e. chaos=40,exa=1).",
	)
}

func validateItems(items []string, errorMsg string) error {
	if containsDuplicate(items) {
		return errors.New(errorMsg + "duplicate items")
	}

	supportedItems := viper.GetStringMapString("bulkItems")
	for _, itemId := range items {
		if _, ok := supportedItems[itemId]; !ok {
			return errors.New(errorMsg + itemId + " is not a supported item")
		}
	}

	return nil
}

func containsDuplicate(array []string) bool {
	set := make(map[string]bool)
	for _, el := range array {
		if _, ok := set[el]; ok {
			return true
		}
		set[el] = true
	}
	return false
}

func getLeague(config Config) string {
	if config.Hardcore {
		return "Hardcore " + config.League
	} else {
		return config.League
	}
}

func analyzeBulkTrades(items []string, capital map[string]int, config Config) error {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
	exchangeClient := api.NewClient(httpClient, getLeague(config))
	tradingPaths := strategy.NewTradingPaths(capital)

	for initialIndex, initialItem := range items {
		for currIndex, currItem := range items {
			if currIndex == initialIndex {
				continue
			}

			bulkTrades, err := exchangeClient.GetBulkTrades(initialItem, currItem, 1)
			if err != nil {
				fmt.Println("Unable to fetch bulk trades:", err)
				return err
			}

			tradeDetails, err := exchangeClient.GetTradeDetails(
				bulkTrades.Id,
				utils.Limit(bulkTrades.TradeIds, 20),
			)
			if err != nil {
				fmt.Println("Unable to fetch trade details:", initialItem, currItem)
				return err
			}

			tradeDetails = filterTradeDetails(tradeDetails, config)
			sortTrades(tradeDetails, config)
			if len(*tradeDetails) > 0 {
				if err := tradingPaths.Set(initialItem, currItem, tradeDetails); err != nil {
					fmt.Println(err)
					return err
				}
			}
		}
	}

	if err := tradingPaths.Analyze(); err != nil {
		fmt.Println("Unable to analyze bulk trades:", err)
		return err
	}

	return nil
}

func filterTradeDetails(tradeDetails *[]api.TradeDetail, config Config) *[]api.TradeDetail {
	filteredTrades := make([]api.TradeDetail, 0, len(*tradeDetails))
	for _, trade := range *tradeDetails {
		if config.ExcludeAFK && trade.AFK {
			continue
		}
		if utils.Contains(config.IgnoredPlayers, trade.Account) {
			continue
		}
		filteredTrades = append(filteredTrades, trade)
	}
	return &filteredTrades
}

func sortTrades(tradeDetails *[]api.TradeDetail, config Config) {
	hasFavorite := len(config.FavoritePlayers) != 0
	less := func(i, j int) bool {
		curr := (*tradeDetails)[i]
		next := (*tradeDetails)[j]
		// Prefer trades with highest value
		if next.Ratio < curr.Ratio {
			return true
		} else if next.Ratio == curr.Ratio {
			currFavorite := utils.Contains(config.FavoritePlayers, curr.Account)
			nextFavorite := utils.Contains(config.FavoritePlayers, next.Account)
			// Prefer trades with favorite players
			if hasFavorite && currFavorite && !nextFavorite {
				return true
			} else if currFavorite == nextFavorite {
				// Prefer trades with more stock
				if next.Stock < curr.Stock {
					return true
				}
			}
		}
		return false
	}
	sort.SliceStable(*tradeDetails, less)
}
