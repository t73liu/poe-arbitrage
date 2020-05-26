package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const baseUrl = "https://www.pathofexile.com/api/trade/"

type Trades struct {
	Id       string   `json:"id"`
	TradeIds []string `json:"result"`
	Total    uint16   `json:"total"`
}

type TradeDetail struct {
	Id      string  `json:"id"`
	Item    Item    `json:"item"`
	Listing Listing `json:"listing"`
}

type Item struct {
	Identified bool   `json:"identified"`
	Verified   bool   `json:"verified"`
	Corrupted  bool   `json:"corrupted"`
	Level      uint8  `json:"ilvl"`
	Note       string `json:"note"`
}

type Listing struct {
	Price   Price   `json:"price"`
	Account Account `json:"account"`
	Whisper string  `json:"whisper"`
}

type Price struct {
	Exchange struct {
		Currency string `json:"currency"`
		Amount   uint16 `json:"amount"`
	} `json:"exchange"`
	Item struct {
		Currency string `json:"currency"`
		Amount   uint16 `json:"amount"`
		Stock    uint16 `json:"stock"`
	} `json:"item"`
}

type Account struct {
	Name   string `json:"name"`
	Online struct {
		League string `json:"league"`
		Status string `json:"status"`
	} `json:"online"`
}

type Client struct {
	client http.Client
	league string
}

func NewClient(httpClient http.Client, league string) *Client {
	return &Client{
		client: httpClient,
		league: league,
	}
}

func (c *Client) GetBulkTrades(initialItem, targetItem string, minStock uint16) (*Trades, error) {
	postStr := getPostParams(initialItem, targetItem, minStock)
	req, err := http.NewRequest(
		"POST",
		baseUrl+"exchange/"+c.league,
		bytes.NewBufferString(postStr),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unable to fetch bulk trade details", resp.Status)
		return nil, err
	}

	var bulkTrades Trades
	if err := json.NewDecoder(resp.Body).Decode(&bulkTrades); err != nil {
		return nil, err
	}

	return &bulkTrades, nil
}

// TODO simplify data model
func (c *Client) GetTradeDetails(queryId string, tradeIds []string) (*[]TradeDetail, error) {
	var tradeDetails []TradeDetail
	if len(tradeIds) == 0 {
		return &tradeDetails, nil
	}
	if len(tradeIds) > 20 {
		return nil, errors.New("bulk trade API has a max limit of 20 ids")
	}

	tradeIdsStr := strings.Join(tradeIds, ",")
	req, err := http.NewRequest("GET", baseUrl+"fetch/"+tradeIdsStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	queryParams := req.URL.Query()
	queryParams.Add("exchange", "")
	queryParams.Add("query", queryId)
	req.URL.RawQuery = queryParams.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unable to fetch trades", resp.Status)
		return nil, err
	}

	var tradesResponse map[string][]TradeDetail
	if err := json.NewDecoder(resp.Body).Decode(&tradesResponse); err != nil {
		return nil, err
	}

	result, ok := tradesResponse["result"]
	if ok {
		return &result, nil
	} else {
		return &tradeDetails, nil
	}
}

func getPostParams(initialItem, targetItem string, minStock uint16) string {
	return fmt.Sprintf(
		`{"exchange":{"status":{"option":"online"},"have":["%s"],"want":["%s"],"minimum":%d}}`,
		initialItem,
		targetItem,
		minStock,
	)
}
