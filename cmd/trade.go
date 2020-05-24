package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"poe-arbitrage/api"
	"time"
)

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Check for trading opportunities for bulk items.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("provide at least 2 items")
		}

		if err := validateItems(args, "Invalid arguments: "); err != nil {
			return err
		}

		initialCapital, err := cmd.Flags().GetStringToInt("capital")
		if err != nil {
			fmt.Println("Could not parse --capital argument", err)
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
	Run: func(cmd *cobra.Command, items []string) {
		initialCapital, err := cmd.Flags().GetStringToInt("capital")
		if err != nil {
			fmt.Println("Could not parse --capital argument", err)
			os.Exit(1)
		}

		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println("Failed to parse config", err)
			os.Exit(1)
		}

		analyzeBulkTrades(items, initialCapital, config)
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

func analyzeBulkTrades(items []string, capital map[string]int, config Config) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
	exchangeClient := api.NewClient(httpClient, getLeague(config))

	for currIndex, initialItem := range items {
		for targetIndex := currIndex + 1; targetIndex < len(items); targetIndex++ {
			targetItem := items[targetIndex]
			_, err := exchangeClient.GetBulkTrades(initialItem, targetItem, 1)
			if err != nil {
				fmt.Println("Unable to fetch bulk trades", err)
				os.Exit(1)
			}
			//fmt.Println(trades)
		}
	}
}
