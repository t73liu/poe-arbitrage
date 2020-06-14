package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
)

const baseUrl = "https://www.pathofexile.com/api/trade/"

type Trades struct {
	Id       string   `json:"id"`
	TradeIds []string `json:"result"`
	Total    uint     `json:"total"`
}

// GGG trade detail is not exposed since it is very verbose and subject to change
type tradeDetail struct {
	Id   string `json:"id"`
	Item struct {
		Identified bool   `json:"identified"`
		Verified   bool   `json:"verified"`
		Corrupted  bool   `json:"corrupted"`
		Level      uint8  `json:"ilvl"`
		Note       string `json:"note"`
	} `json:"item"`
	Listing struct {
		Price struct {
			Exchange struct {
				Currency string  `json:"currency"`
				Amount   float64 `json:"amount"`
			} `json:"exchange"`
			Item struct {
				Currency string  `json:"currency"`
				Amount   float64 `json:"amount"`
				Stock    uint    `json:"stock"`
			} `json:"item"`
		} `json:"price"`
		Account struct {
			Name   string `json:"name"`
			Online struct {
				League string `json:"league"`
				Status string `json:"status"`
			} `json:"online"`
		} `json:"account"`
		Whisper string `json:"whisper"`
	} `json:"listing"`
}

type TradeDetail struct {
	Account     string
	AFK         bool
	Whisper     string
	PriceAmount uint
	PriceUnit   string
	ItemAmount  uint
	ItemUnit    string
	Stock       uint
	Ratio       float64
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

func (c *Client) GetBulkTrades(initialItem, targetItem string, minStock uint) (*Trades, error) {
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

func (c *Client) getTradeDetails(queryId string, tradeIds []string) (*[]tradeDetail, error) {
	var tradeDetails []tradeDetail
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

	var tradesResponse map[string][]tradeDetail
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

func (c *Client) GetTradeDetails(queryId string, tradeIds []string) (*[]TradeDetail, error) {
	tradeDetails, err := c.getTradeDetails(queryId, tradeIds)
	if err != nil {
		return nil, err
	}

	formattedTradeDetails := make([]TradeDetail, 0, len(*tradeDetails))
	for _, tradeDetail := range *tradeDetails {
		cost := tradeDetail.Listing.Price.Exchange.Amount
		itemAmount := tradeDetail.Listing.Price.Item.Amount
		// Rounding partial amounts can increase trade costs but this rarely happens
		roundedPriceAmount := uint(math.Ceil(cost))
		roundedItemAmount := uint(math.Floor(itemAmount))
		formattedTrade := TradeDetail{
			Account:     tradeDetail.Listing.Account.Name,
			AFK:         tradeDetail.Listing.Account.Online.Status == "afk",
			Whisper:     tradeDetail.Listing.Whisper,
			PriceAmount: roundedPriceAmount,
			PriceUnit:   tradeDetail.Listing.Price.Exchange.Currency,
			ItemAmount:  roundedItemAmount,
			ItemUnit:    tradeDetail.Listing.Price.Item.Currency,
			Ratio:       itemAmount / cost,
			Stock:       tradeDetail.Listing.Price.Item.Stock,
		}
		if roundedItemAmount >= 1 && roundedPriceAmount >= 1 {
			formattedTradeDetails = append(formattedTradeDetails, formattedTrade)
		}
	}
	return &formattedTradeDetails, nil
}

func getPostParams(initialItem, targetItem string, minStock uint) string {
	return fmt.Sprintf(
		`{"exchange":{"status":{"option":"online"},"have":["%s"],"want":["%s"],"minimum":%d}}`,
		initialItem,
		targetItem,
		minStock,
	)
}
