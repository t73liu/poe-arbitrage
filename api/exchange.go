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

const baseURL = "https://www.pathofexile.com/api/trade/"

type Trades struct {
	ID       string   `json:"id"`
	TradeIDs []string `json:"result"`
	Total    uint     `json:"total"`
}

// GGG trade detail is not exposed since it is very verbose and subject to change
type tradeDetail struct {
	ID   string `json:"id"`
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
		baseURL+"exchange/"+c.league,
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
		return nil, fmt.Errorf("request failed with %s", resp.Status)
	}

	var bulkTrades Trades
	if err := json.NewDecoder(resp.Body).Decode(&bulkTrades); err != nil {
		return nil, err
	}

	return &bulkTrades, nil
}

func (c *Client) getTradeDetails(queryID string, tradeIDs []string) (*[]tradeDetail, error) {
	var tradeDetails []tradeDetail
	if len(tradeIDs) == 0 {
		return &tradeDetails, nil
	}
	if len(tradeIDs) > 20 {
		return nil, errors.New("bulk trade API has a max limit of 20 ids")
	}

	tradeIDsStr := strings.Join(tradeIDs, ",")
	req, err := http.NewRequest("GET", baseURL+"fetch/"+tradeIDsStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	queryParams := req.URL.Query()
	queryParams.Add("exchange", "")
	queryParams.Add("query", queryID)
	req.URL.RawQuery = queryParams.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with %s", resp.Status)
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

func (c *Client) GetTradeDetails(queryID string, tradeIDs []string) (*[]TradeDetail, error) {
	tradeDetails, err := c.getTradeDetails(queryID, tradeIDs)
	if err != nil {
		return nil, err
	}

	formattedTradeDetails := make([]TradeDetail, 0, len(*tradeDetails))
	for _, tradeDetail := range *tradeDetails {
		cost := tradeDetail.Listing.Price.Exchange.Amount
		itemAmount := tradeDetail.Listing.Price.Item.Amount
		// TODO: Rounding up partial amounts will increase trade costs but this
		// rarely happens
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
