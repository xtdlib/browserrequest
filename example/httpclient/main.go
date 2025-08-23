package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/xtdlib/browserrequest"
)

type DualAssetsOrdersRequest struct {
	ProductType        int    `json:"product_type"`
	OnlyEffectiveOrder bool   `json:"only_effective_order"`
	StartAt            *int64 `json:"start_at"`
	EndAt              *int64 `json:"end_at"`
	BaseCoin           int    `json:"base_coin"`
	Limit              int    `json:"limit"`
	NextOffsetID       string `json:"next_offset_id"`
}

var reqBody = DualAssetsOrdersRequest{
	ProductType:        2,
	OnlyEffectiveOrder: false,
	StartAt:            nil, // null value
	EndAt:              nil, // null value
	BaseCoin:           858,
	Limit:              10,
}

func main() {
	log.Println("start")
	client := browserrequest.NewHTTPClient(http.DefaultTransport, &browserrequest.Options{
		DebugURL: "ws://localhost:9222",
	})

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://www.bybit.com/x-api/s1/byfi/dual-assets/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
}
