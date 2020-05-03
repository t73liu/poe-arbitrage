package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

const baseUrl = "https://www.pathofexile.com/api/trade/"

type Client struct {
	client     http.Client
	leagueName string
}

type TradeInfo struct {
	Id       string   `json:"id"`
	TradeIds []string `json:"result"`
	Total    uint8    `json:"total"`
}

func NewClient(httpClient http.Client, hardcore bool) *Client {
	leagueName := "Delirium"
	if hardcore {
		leagueName = "Hardcore " + leagueName
	}
	return &Client{
		client:     httpClient,
		leagueName: leagueName,
	}
}

// TODO: error handling, use currencies, return trades
func (c *Client) GetCurrencyTrades(initialCurrency, targetCurrency string) (*TradeInfo, error) {
	req, err := http.NewRequest("POST", baseUrl+"exchange/"+c.leagueName, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	var tradeInfo TradeInfo
	if err := json.NewDecoder(resp.Body).Decode(&tradeInfo); err != nil {
		return nil, err
	}
	if tradeInfo.Total > 0 {
		tradeIdsStr := strings.Join(tradeInfo.TradeIds, ",")
		tradesReq, err := http.NewRequest("GET", baseUrl+"fetch/"+tradeIdsStr, nil)
		if err != nil {
			return nil, err
		}
		queryParams := tradesReq.URL.Query()
		queryParams.Add("query", tradeInfo.Id)
		queryParams.Add("exchange", "")
	}
	return &tradeInfo, nil
}
