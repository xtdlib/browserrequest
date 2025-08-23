package main

import (
	"bytes"
	"context"
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

func main() {
	cookies := browserrequest.NewSession(&browserrequest.Options{
		DebugURL: "ws://localhost:9222",
	})

	reqBody := DualAssetsOrdersRequest{
		ProductType:        2,
		OnlyEffectiveOrder: false,
		StartAt:            nil, // null value
		EndAt:              nil, // null value
		BaseCoin:           858,
		Limit:              10,
		NextOffsetID:       "2110673.2",
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "https://www.bybit.com/x-api/s1/byfi/dual-assets/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("accept", "application/json")

	ctx := context.Background()
	if err := cookies.SetRequest(ctx, req); err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
}
