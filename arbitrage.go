package main

import (
	"log"
	"net/http"
	"os"
	"poe-arbitrage/api"
	"time"
)

// TODO: analyze trades according to CLI options
func analyzeTrades(trades string, options string) {

}

// TODO: parse options, read currencies.json, validations
func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	client := http.Client{Timeout: 15 * time.Second}
	exchangeClient := api.NewClient(client, false)

	trades, err := exchangeClient.GetCurrencyTrades("exa", "chaos")
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println(trades)
}
