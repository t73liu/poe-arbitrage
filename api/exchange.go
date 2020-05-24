package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const baseUrl = "https://www.pathofexile.com/api/trade/"

type BulkTrades struct {
	Id       string   `json:"id"`
	TradeIds []string `json:"result"`
	Total    uint16   `json:"total"`
}

type Trades struct {
	Result []Trade `json:"result"`
}

type Trade struct {
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
		Currency string
		Amount   uint16
	}
	Item struct {
		Currency string
		Amount   uint16
		Stock    uint16
	}
}

type Account struct {
	Name   string       `json:"name"`
	Online OnlineStatus `json:"online"`
}

type OnlineStatus struct {
	League string `json:"league"`
	Status string `json:"status"`
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

func (c *Client) GetBulkTrades(initialItem, targetItem string, minStock uint16) (*BulkTrades, error) {
	postStr := getPostParams(initialItem, targetItem, minStock)
	bulkReq, err := http.NewRequest(
		"POST",
		baseUrl+"exchange/"+c.league,
		bytes.NewBufferString(postStr),
	)
	if err != nil {
		return nil, err
	}

	bulkReq.Header.Set("Content-Type", "application/json")
	bulkResp, err := c.client.Do(bulkReq)
	if err != nil {
		return nil, err
	}
	defer bulkResp.Body.Close()

	if bulkResp.StatusCode != http.StatusOK {
		fmt.Println("Unable to fetch bulk trade details", bulkResp.Status)
		os.Exit(1)
	}

	var bulkTrades BulkTrades
	if err := json.NewDecoder(bulkResp.Body).Decode(&bulkTrades); err != nil {
		return nil, err
	}

	if bulkTrades.Total > 0 {
		// Bulk Trade API has a max limit of 20 ids
		tradeIds := limitStringArray(bulkTrades.TradeIds, 20)
		tradeIdsStr := strings.Join(tradeIds, ",")

		tradesReq, err := http.NewRequest("GET", baseUrl+"fetch/"+tradeIdsStr, nil)
		if err != nil {
			return nil, err
		}
		tradesReq.Header.Set("Content-Type", "application/json")
		queryParams := tradesReq.URL.Query()
		queryParams.Add("exchange", "")
		queryParams.Add("query", bulkTrades.Id)
		tradesReq.URL.RawQuery = queryParams.Encode()

		tradesResp, err := c.client.Do(tradesReq)
		if err != nil {
			return nil, err
		}
		defer tradesResp.Body.Close()

		if tradesResp.StatusCode != http.StatusOK {
			fmt.Println("Unable to fetch trades", tradesResp.Status)
			os.Exit(1)
		}

		//if err := json.NewDecoder(tradesResp.Body).Decode(&trades); err != nil {
		// return nil, err
		//}
	}
	return &bulkTrades, nil
}

func getPostParams(initialItem, targetItem string, minStock uint16) string {
	return fmt.Sprintf(
		`{"exchange":{"status":{"option":"online"},"have":["%s"],"want":["%s"],"minimum":%d}}`,
		initialItem,
		targetItem,
		minStock,
	)
}

func limitStringArray(array []string, maxSize int) []string {
	if len(array) <= maxSize {
		return array
	}
	result := make([]string, maxSize)
	for i := 0; i < maxSize; i++ {
		result[i] = array[i]
	}
	return result
}
