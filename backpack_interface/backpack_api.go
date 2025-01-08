package main // 声明 main 包，表明当前是一个可执行程序

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var baseURL = "https://api.backpack.exchange"

var client = &http.Client{
	Timeout: 6 * time.Second,
}

type Key struct {
	apiKey string
	secret string
}

// k线
type Klines struct {
	symbol    string
	interval  string
	startTime int
	endTime   int
}

// 获取历史交易
type HistoryTrades struct {
	symbol string
	limit  int
	offset int
}

// 存取历史
type DepositeHistory struct {
	key    Key
	limit  int
	offset int
}

// 请求提取
type RequestWithdraw struct {
	key            Key
	address        string
	blockchain     string
	clientId       string
	quantity       string
	symbol         string
	twoFactorToken string
	subaccountId   string
}

// 订单历史
type OrderHistory struct {
	key     Key
	orderId string
	symbol  string
	offset  int
	limit   int
}

// 填充的订单历史
type FillHistory struct {
	key     Key
	orderId string
	from    int
	to      int
	symbol  string
	limit   int
	offset  int
}

// 提取历史记录
type WithdrawHistory struct {
	key    Key
	limit  int
	offset int
}

// 打开订单记录
type OpenOrder struct {
	key      Key
	clientId int
	orderId  string
	symbol   string
}

// 创建订单
type CreateOrder struct {
	key                 Key
	clientId            string
	orderType           string
	postOnly            bool
	price               float64
	quantity            float64
	quoteQuantity       float64
	selfTradePrevention string
	side                string
	symbol              string
	timeInForce         string
	triggerPrice        float64
}

// 取消打开的订单
type CancelTokenOrder struct {
	key      Key
	clientId int
	orderId  string
	symbol   string
}

func convertMap(mapInterface map[string]interface{}) map[string]string {
	mapString := make(map[string]string)

	for key, value := range mapInterface {
		switch v := value.(type) {
		case string:
			mapString[key] = v
		case int:
			mapString[key] = strconv.Itoa(v)
		// 添加更多的类型转换，根据需要
		default:
			mapString[key] = fmt.Sprintf("%v", value)
		}
	}

	return mapString
}

func getRequest(
	url string,
	head map[string]string,
	params map[string]interface{},
) (map[string]interface{}, error) {
	reqURL := baseURL + url

	paramString := convertMap(params)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		panic(err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range head {
		req.Header.Set(k, v)
	}

	// 设置查询参数
	q := req.URL.Query()
	for k, v := range paramString {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 解码响应体
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	return result, nil
}

func postRequest(
	url string,
	head map[string]string,
	proxy map[string]string,
	params map[string]interface{},
) (map[string]interface{}, error) {
	reqURL := baseURL + url

	jsonParams, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonParams))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range head {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	return result, nil
}

func deleteRequest(
	url string,
	head map[string]string,
	proxy map[string]string,
	params map[string]interface{},
) (map[string]interface{}, error) {
	reqURL := baseURL + url

	jsonParams, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("DELETE", reqURL, bytes.NewBuffer(jsonParams))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range head {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	return result, nil
}

// 检索交易所支持的所有资产。
func getAssets() ([]map[string]interface{}, error) {
	url := "/api/v1/assets"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}

	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make([]map[string]interface{}, len(data))
	for i, _ := range data {
		v, _ := data[i].(map[string]interface{})
		tokensData, _ := v["tokens"].([]interface{})

		tokens := make([]map[string]interface{}, len(v["tokens"].([]interface{})))
		for j, token := range tokensData {
			tokenInfo := token.(map[string]interface{})
			tokens[j] = map[string]interface{}{
				"blockchain":        tokenInfo["blockchain"],
				"depositEnabled":    tokenInfo["depositEnabled"],
				"minimumDeposit":    tokenInfo["minimumDeposit"],
				"withdrawEnabled":   tokenInfo["withdrawEnabled"],
				"minimumWithdrawal": tokenInfo["minimumWithdrawal"],
				"maximumWithdrawal": tokenInfo["maximumWithdrawal"],
				"withdrawalFee":     tokenInfo["withdrawalFee"],
			}
		}
		// 将字符串转换为整数
		num, err := strconv.Atoi(i)
		if err != nil {
			fmt.Println("转换失败:")
			panic(err)
		}
		rst[num] = map[string]interface{}{
			"symbol": v["symbol"],
			"tokens": tokens,
		}
	}
	fmt.Println("Assets:", rst)
	return rst, nil
}

// 检索交易所支持的所有市场。
func getMarkets() ([]map[string]interface{}, error) {
	url := "/api/v1/markets"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}

	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}
	rst := make([]map[string]interface{}, len(data))
	for i, item := range data {
		v := item.(map[string]interface{})
		filters := v["filters"].(map[string]interface{})
		// 将字符串转换为整数
		num, err := strconv.Atoi(i)
		if err != nil {
			fmt.Println("转换失败:", err)
		}
		rst[num] = map[string]interface{}{
			"symbol":      v["symbol"],
			"baseSymbol":  v["baseSymbol"],
			"quoteSymbol": v["quoteSymbol"],
			"filters": map[string]interface{}{
				"price": map[string]interface{}{
					"minPrice": filters["minPrice"],
					"maxPrice": filters["maxPrice"],
					"tickSize": filters["tickSize"],
				},
				"quantity": map[string]interface{}{
					"minQuantity": filters["minQuantity"],
					"maxQuantity": filters["maxQuantity"],
					"stepSize":    filters["stepSize"],
				},
				"leverage": map[string]interface{}{
					"minLeverage": filters["minLeverage"],
					"maxLeverage": filters["maxLeverage"],
					"stepSize":    filters["stepSize"],
				},
			},
		}
	}
	fmt.Println("Markets:", rst)
	return rst, nil
}

// 检索过去24小时内给定市场代码的汇总统计数据。
func getTicker(symbol string) (map[string]interface{}, error) {
	url := "/api/v1/ticker"
	params := map[string]interface{}{"symbol": symbol}
	data, err := getRequest(url, nil, params)
	if err != nil {
		panic(err)
	}
	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println("Ticker:", data)
	return data, nil
}

// 检索过去24小时所有市场代码的汇总统计数据。
func getTickers() ([]map[string]interface{}, error) {
	url := "/api/v1/tickers"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}
	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}

	var rst []map[string]interface{}
	for _, values := range data {
		v := values.(map[string]interface{})
		ticker := map[string]interface{}{
			"symbol":             v["symbol"],
			"firstPrice":         v["firstPrice"],
			"lastPrice":          v["lastPrice"],
			"priceChange":        v["priceChange"],
			"priceChangePercent": v["priceChangePercent"],
			"high":               v["high"],
			"low":                v["low"],
			"volume":             v["volume"],
			"quoteVolume":        v["quoteVolume"],
			"trades":             v["trades"],
		}
		rst = append(rst, ticker)
	}

	fmt.Println("Tickers:", rst)
	return rst, nil
}

// 检索给定市场符号的订单深度。
func getDepth(symbol string) (map[string]interface{}, error) {
	url := "/api/v1/depth"
	params := map[string]interface{}{"symbol": symbol}
	data, err := getRequest(url, nil, params)
	if err != nil {
		panic(err)
	}
	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println(data)
	return data, nil
}

// 获取给定市场代码的k线
func getKLines(
	k Klines,
) ([]map[string]interface{}, error) {
	url := "/api/v1/klines"
	params := map[string]interface{}{
		"symbol":    k.symbol,
		"interval":  k.interval,
		"startTime": k.startTime,
		"endTime":   k.endTime,
	}

	data, err := getRequest(url, nil, params)
	if err != nil {
		panic(err)
	}

	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}

	var kLines []map[string]interface{}
	for _, values := range data {
		v := values.(map[string]interface{})
		kLine := map[string]interface{}{
			"start":  v["start"],
			"open":   v["open"],
			"high":   v["high"],
			"low":    v["low"],
			"close":  v["close"],
			"end":    v["end"],
			"volume": v["volume"],
			"trades": v["trades"],
		}
		kLines = append(kLines, kLine)
	}

	return kLines, nil
}

/** ********************************** system ******************************** */
// 得到系统状态
func getStatus() (map[string]interface{}, error) {
	url := "/api/v1/status"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	return data, err
}

// 得到ping
func getPing() (map[string]interface{}, error) {
	url := "/api/v1/ping"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	return data, err
}

// 得到当前系统时间
func getSystemTime() (map[string]interface{}, error) {
	url := "/api/v1/time"
	data, err := getRequest(url, nil, nil)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	return data, err
}

/** ********************************** 获取交易信息 ******************************** */
// 获取最近的交易
func getRecentTrades(symbol string, limit int) ([]map[string]interface{}, error) {
	url := "/api/v1/trades"
	params := map[string]interface{}{
		"symbol": symbol,
		"limit":  limit,
	}
	data, err := getRequest(url, nil, params)
	if err != nil {
		panic(err)
	}

	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"id":            v["id"].(int64),
			"price":         v["price"].(string),
			"quantity":      v["quantity"].(string),
			"quoteQuantity": v["quoteQuantity"].(string),
			"timestamp":     v["timestamp"].(int64),
			"isBuyerMaker":  v["isBuyerMaker"].(bool),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("Recent Trades:", rst)
	return rst, nil
}

// 获取历史交易
func getHistoricalTrades(
	h HistoryTrades,
) ([]map[string]interface{}, error) {
	url := "/api/v1/trades/history"
	params := map[string]interface{}{
		"symbol": h.symbol,
		"limit":  h.limit,
		"offset": h.limit,
	}
	data, err := getRequest(url, nil, params)
	if err != nil {
		panic(err)
	}
	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"id":            v["id"].(int64),
			"price":         v["price"].(string),
			"quantity":      v["quantity"].(string),
			"quoteQuantity": v["quoteQuantity"].(string),
			"timestamp":     v["timestamp"].(int64),
			"isBuyerMaker":  v["isBuyerMaker"].(bool),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("History Trades:", rst)
	return rst, nil
}

// capital
func generateSignature(
	instruction,
	apikey,
	secret string,
	payload map[string]interface{},
) map[string]string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	window := "10000"

	payloadString := convertMap(payload)

	data := make(map[string]string)
	data["instruction"] = instruction
	for k, v := range payloadString {
		data[k] = v
	}
	data["timestamp"] = timestamp
	data["window"] = window

	// Sort data by keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sortedData []string
	for _, k := range keys {
		sortedData = append(sortedData, k+"="+data[k])
	}

	// Create a string to sign
	signString := strings.Join(sortedData, "&")

	// Sign the string using the private key
	privateKeyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		panic(err)
	}
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	signature := ed25519.Sign(privateKey, []byte(signString))

	// Encode the signature in base64
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	head := map[string]string{
		"X-API-Key":   apikey,
		"X-Signature": signatureB64,
		"X-Timestamp": timestamp,
		"X-Window":    window,
	}

	return head
}

// 获取账户余额和余额状态
func getBalances(
	key Key,
) (map[string]interface{}, error) {
	url := "/api/v1/capital"

	headers := generateSignature("balanceQuery", key.apiKey, key.secret, nil)

	data, err := getRequest(url, headers, nil)
	if err != nil {
		panic(err)
	}

	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make(map[string]interface{})
	for asset, values := range data {
		v := values.(map[string]interface{})
		available, _ := strconv.ParseFloat(v["available"].(string), 64)
		locked, _ := strconv.ParseFloat(v["locked"].(string), 64)
		staked, _ := strconv.ParseFloat(v["staked"].(string), 64)
		total := available + locked + staked
		assetInfo := map[string]interface{}{
			"available": available,
			"locked":    locked,
			"staked":    staked,
			"total":     total,
		}
		rst[asset] = assetInfo
	}
	fmt.Println("User Balances:", rst)
	return rst, nil
}

// 获取存款历史记录
func getDepositeHistory(
	DE DepositeHistory,
) ([]map[string]interface{}, error) {
	url := "/wapi/v1/capital/deposits"
	params := map[string]interface{}{
		"limit":  DE.limit,
		"offset": DE.offset,
	}
	headers := generateSignature("depositQueryAll", DE.key.apiKey, DE.key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"id":                      v["id"].(int32),
			"toAddress":               v["toAddress"].(string),
			"fromAddress":             v["fromAddress"].(string),
			"confirmationBlockNumber": v["confirmationBlockNumber"].(int64),
			"providerId":              v["providerId"].(string),
			"source":                  v["source"].(string),
			"status":                  v["status"].(string),
			"transactionHash":         v["transactionHash"].(string),
			"subaccountId":            v["subaccountId"].(int32),
			"symbol":                  v["symbol"].(string),
			"quantity":                v["quantity"].(string),
			"createdAt":               v["createdAt"].(string),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("Deposite History:", rst)
	return rst, nil
}

// 获取存款地址
func getDepositorAddress(
	key Key,
	blockchain string,
) (map[string]interface{}, error) {
	url := "/wapi/v1/capital/deposit/address"
	params := map[string]interface{}{"blockchain": blockchain}
	headers := generateSignature("depositAddressQuery", key.apiKey, key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println("Depositor Address:", data)
	return data, err
}

// 获取提款历史记录
func getWithdrawHistory(
	wh WithdrawHistory,
) ([]map[string]interface{}, error) {
	url := "/wapi/v1/capital/withdrawals"
	params := map[string]interface{}{
		"limit":  wh.limit,
		"offset": wh.limit,
	}
	headers := generateSignature("withdrawalQueryAll", wh.key.apiKey, wh.key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if data != nil && data["status"] == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"id":              v["id"].(int32),
			"blockchain":      v["blockchain"].(string),
			"clientId":        v["clientId"].(string),
			"identifier":      v["identifier"].(string),
			"quantity":        v["quantity"].(string),
			"fee":             v["fee"].(string),
			"symbol":          v["symbol"].(string),
			"status":          v["status"].(string),
			"subaccountId":    v["subaccountId"].(int32),
			"toAddress":       v["toAddress"].(string),
			"transactionHash": v["transactionHash"].(string),
			"createdAt":       v["createdAt"].(string),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("Withdraw History:", rst)
	return rst, nil
}

// 请求提款
func requestWithdrawal(
	rw RequestWithdraw,
) (map[string]interface{}, error) {
	url := "/wapi/v1/capital/withdrawals"
	params := map[string]interface{}{
		"address":        rw.address,
		"blockchain":     rw.blockchain,
		"clientId":       rw.clientId,
		"quantity":       rw.quantity,
		"symbol":         rw.symbol,
		"twoFactorToken": rw.twoFactorToken,
	}
	head := generateSignature("withdraw", rw.key.apiKey, rw.key.secret, params)
	data, err := postRequest(url, head, nil, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println("Request Withdraw:", data)
	return data, nil
}

/** ********************************** 订单历史记录 ******************************** */
// 获取订单历史记录。
func getOrderHistory(
	oh OrderHistory,
) ([]map[string]interface{}, error) {
	url := "/wapi/v1/history/orders"
	params := map[string]interface{}{
		"orderId": oh.orderId,
		"symbol":  oh.symbol,
		"limit":   oh.limit,
		"offset":  oh.offset,
	}

	headers := generateSignature("orderHistoryQueryAll", oh.key.apiKey, oh.key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"id":                  v["id"].(string),
			"orderType":           v["orderType"].(string),
			"symbol":              v["symbol"].(string),
			"side":                v["side"].(string),
			"price":               v["price"].(string),
			"triggerPrice":        v["triggerPrice"].(string),
			"quantity":            v["quantity"].(string),
			"quoteQuantity":       v["quoteQuantity"].(string),
			"timeInForce":         v["timeInForce"].(int32),
			"selfTradePrevention": v["selfTradePrevention"].(string),
			"postOnly":            v["postOnly"].(string),
			"status":              v["status"].(string),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("Order History:", rst)
	return rst, nil
}

// 获取填充订单历史记录
func getFillHistory(
	FH FillHistory,
) (map[string]interface{}, error) {
	url := "/wapi/v1/history/fills"
	params := map[string]interface{}{
		"orderId": FH.orderId,
		"from":    FH.from,
		"to":      FH.to,
		"symbol":  FH.symbol,
		"limit":   FH.limit,
		"offset":  FH.limit,
	}
	headers := generateSignature("fillHistoryQueryAll", FH.key.apiKey, FH.key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}

	rst := make(map[string]interface{})
	for asset, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"tradeId":   v["tradeId"].(int64),
			"orderId":   v["orderType"].(string),
			"symbol":    v["symbol"].(string),
			"side":      v["side"].(string),
			"price":     v["price"].(string),
			"quantity":  v["quantity"].(string),
			"fee":       v["fee"].(string),
			"feeSymbol": v["feeSymbol"].(string),
			"isMaker":   v["isMaker"].(int32),
			"timestamp": v["timestamp"].(string),
		}
		rst[asset] = assetInfo
	}
	fmt.Println("Fill History:", rst)
	return rst, nil
}

/** ********************************** 订单相关 ******************************** */
//得到开仓的订单
func getTokenOpenOrder(
	o OpenOrder,
) (map[string]interface{}, error) {
	url := "/api/v1/order"
	params := map[string]interface{}{
		"clientId": o.clientId,
		"symbol":   o.symbol,
		"orderId":  o.orderId,
	}
	headers := generateSignature("orderQuery", o.key.apiKey, o.key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println("This Open Order:", data)
	return data, err
}

// 执行订单
func createOrder(
	co CreateOrder,
) (map[string]interface{}, error) {
	url := "/api/v1/order"

	params := map[string]interface{}{
		"clientId":            co.clientId,
		"orderType":           co.orderType,
		"postOnly":            co.postOnly,
		"price":               co.price,
		"quantity":            co.quantity,
		"quoteQuantity":       co.quoteQuantity,
		"selfTradePrevention": co.selfTradePrevention,
		"side":                co.side,
		"symbol":              co.symbol,
		"timeInForce":         co.timeInForce,
		"triggerPrice":        co.triggerPrice,
	}

	head := generateSignature("orderExecute", co.key.apiKey, co.key.secret, params)

	data, err := postRequest(url, head, nil, params)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, err
	}
	fmt.Println("create order:", data)
	return data, nil
}

// 检索某个token所有未结订单
func getTokenOpenAllOrders(
	key Key,
	symbol string,
) ([]map[string]interface{}, error) {
	url := "/api/v1/orders"
	params := map[string]interface{}{"symbol": symbol}
	headers := generateSignature("orderQueryAll", key.apiKey, key.secret, params)
	data, err := getRequest(url, headers, params)
	if err != nil {
		panic(err)
	}

	rst := make([]map[string]interface{}, len(data))
	for _, values := range data {
		v := values.(map[string]interface{})
		assetInfo := map[string]interface{}{
			"orderType":             v["orderType"].(string),
			"id":                    v["id"].(string),
			"clientId":              v["clientId"].(uint32),
			"symbol":                v["symbol"].(string),
			"side":                  v["side"].(string),
			"quantity":              v["quantity"].(string),
			"executedQuantity":      v["executedQuantity"].(string),
			"quoteQuantity":         v["quoteQuantity"].(string),
			"executedQuoteQuantity": v["executedQuoteQuantity"].(string),
			"triggerPrice":          v["triggerPrice"].(string),
			"timeInForce":           v["timeInForce"].(string),
			"selfTradePrevention":   v["selfTradePrevention"].(string),
			"status":                v["status"].(string),
			"createdAt":             v["createdAt"].(int64),
		}
		rst = append(rst, assetInfo)
	}
	fmt.Println("Token All OpenOrders:", rst)
	return rst, nil
}

// 从订单簿中取消某个未结订单。
func cancelOpenOrder(
	CTO CancelTokenOrder,
) (map[string]interface{}, error) {
	url := "/api/v1/order"
	params := map[string]interface{}{
		"clientId": CTO.clientId,
		"orderId":  CTO.orderId,
		"symbol":   CTO.symbol,
	}
	head := generateSignature("orderCancel", CTO.key.apiKey, CTO.key.secret, params)
	//转换
	paramsInterface := make(map[string]interface{})
	for key, value := range params {
		paramsInterface[key] = value
	}
	data, err := deleteRequest(url, nil, head, paramsInterface)
	if err != nil {
		panic(err)
	}
	if status, ok := data["status"]; ok && status == "fail" {
		fmt.Println(data)
		return nil, nil
	}
	fmt.Println("Cancel This OpenOrder:", data)
	return data, nil
}

// 从订单簿中取消所有未结订单。
func cancelOpenOrders(
	key Key,
	symbol string,
	proxy map[string]string,
) (map[string]interface{}, error) {
	url := "/api/v1/orders"
	params := map[string]interface{}{"symbol": symbol}

	head := generateSignature("orderCancelAll", key.apiKey, key.secret, params)
	//转换
	paramsInterface := make(map[string]interface{})
	for key, value := range params {
		paramsInterface[key] = value
	}

	count := 0
	for {
		if count > 3 {
			return nil, nil
		}
		data, err := deleteRequest(url, proxy, head, paramsInterface)
		if err != nil {
			return nil, err
		}
		if data != nil && data["status"] == "fail" {
			fmt.Println("retry cancel orders")
			count++
		} else {
			return data, nil
		}
	}
}

func main() {
	systemTime, err := getSystemTime()
	if err != nil {
		panic(err)
	}

	// 输出系统时间
	fmt.Println("Assets:", systemTime)
}
