package pkg

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/shopspring/decimal"
	"resty.dev/v3"
)

const url = "https://api.0x.org/swap/allowance-holder/price"

var ox_api_keys_1 = []string{"3b4d5770-e20e-439c-ac15-776a92aeee70", "f9fb0f52-a19a-4f06-8b89-0d938ea1145b"}

// parseTokenAmountWithToken 解析代币数量，根据代币的小数位数进行转换
func parseTokenAmountWithToken(amount string, token string) (decimal.Decimal, error) {
	// 默认小数位数为18
	decimals := int64(18)

	// TODO: 这里可以添加获取代币信息的逻辑
	// tokenInfo := getTokenInfo(strings.ToLower(token))
	// if tokenInfo != nil {
	//     decimals = tokenInfo.Decimals
	// }

	// 解析数量字符串
	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("解析数量失败: %v", err)
	}

	// 计算实际数量：amount / (10^decimals)
	result := amountDecimal.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(decimals)))

	return result, nil
}

func GetPrice() {
	var priceResponse PriceResponse
	client := resty.New()
	client.SetDebug(true)
	_, err := client.R().SetHeaders(map[string]string{
		"0x-api-key":   ox_api_keys_1[rand.Intn(len(ox_api_keys_1))],
		"0x-version":   "v2",
	}).SetQueryParams(map[string]string{
		"chainId": "480",
		// "sellToken":        "0x2cFc85d8E48F8EAB294be644d9E25C3030863003",// 0.5777512565597034
		"sellToken":        "0x79A02482A880bCe3F13E09da970dC34dB4cD24D1",//0.0074
		"buyToken":         "0xED49fE44fD4249A09843C2Ba4bba7e50BECa7113",
		"sellAmount":       "1000000000000",
		"slippageBps":      "100",
		"taker":            "0x4Db902Ce3d3D35Dd1172E21c945356C59A4B9f4a",
		"swapFeeRecipient": "0xC25B58325B09e62990F818aD10fD2a9E3E1888e2",
		"swapFeeBps":       "1",
		"swapFeeToken":     "0xED49fE44fD4249A09843C2Ba4bba7e50BECa7113",
	}).SetResult(&priceResponse).Get(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("response:", priceResponse.LiquidityAvailable)
	fmt.Println("buyToken:", priceResponse.MinBuyAmount)

	// 使用封装的函数处理 MinBuyAmount
	min_buy_amount_decimal, err := parseTokenAmountWithToken(priceResponse.MinBuyAmount, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MinBuyAmount (decimal): %s\n", min_buy_amount_decimal.String())

	// 使用封装的函数处理 SellAmount
	sell_amount_decimal, err := parseTokenAmountWithToken("1000000000000", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SellAmount (decimal): %s\n", sell_amount_decimal.String())

	// 使用封装的函数处理 BuyAmount
	if priceResponse.BuyAmount != "" {
		buy_amount_decimal, err := parseTokenAmountWithToken(priceResponse.BuyAmount, "")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BuyAmount (decimal): %s\n", buy_amount_decimal.String())
	}
}

type PriceResponse struct {
	AllowanceTarget string `json:"allowanceTarget"`
	BlockNumber     string `json:"blockNumber"`
	BuyAmount       string `json:"buyAmount"`
	BuyToken        string `json:"buyToken"`
	Fees            struct {
		IntegratorFee interface{} `json:"integratorFee"`
		ZeroExFee     interface{} `json:"zeroExFee"`
		GasFee        interface{} `json:"gasFee"`
	} `json:"fees"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Issues   struct {
		Allowance struct {
			Actual  string `json:"actual"`
			Spender string `json:"spender"`
		} `json:"allowance"`
		Balance struct {
			Token    string `json:"token"`
			Actual   string `json:"actual"`
			Expected string `json:"expected"`
		} `json:"balance"`
		SimulationIncomplete bool          `json:"simulationIncomplete"`
		InvalidSourcesPassed []interface{} `json:"invalidSourcesPassed"`
	} `json:"issues"`
	LiquidityAvailable bool   `json:"liquidityAvailable"`
	MinBuyAmount       string `json:"minBuyAmount"`
	Route              struct {
		Fills []struct {
			From          string `json:"from"`
			To            string `json:"to"`
			Source        string `json:"source"`
			ProportionBps string `json:"proportionBps"`
		} `json:"fills"`
		Tokens []struct {
			Address string `json:"address"`
			Symbol  string `json:"symbol"`
		} `json:"tokens"`
	} `json:"route"`
	SellAmount    string `json:"sellAmount"`
	SellToken     string `json:"sellToken"`
	TokenMetadata struct {
		BuyToken struct {
			BuyTaxBps  string `json:"buyTaxBps"`
			SellTaxBps string `json:"sellTaxBps"`
		} `json:"buyToken"`
		SellToken struct {
			BuyTaxBps  string `json:"buyTaxBps"`
			SellTaxBps string `json:"sellTaxBps"`
		} `json:"sellToken"`
	} `json:"tokenMetadata"`
	TotalNetworkFee string `json:"totalNetworkFee"`
	Zid             string `json:"zid"`
}
