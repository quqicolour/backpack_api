package main // 声明 main 包，表明当前是一个可执行程序

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"
	"websocket-main"
)

var (
	interrupt = make(chan os.Signal, 1)
	done      = make(chan struct{})
)

// WebSocket 连接细节
const (
	wsURL    = "wss://ws.backpack.exchange"
	readWait = 3 * time.Second
)

// OrderUpdate表示订单更新事件的结构
type OrderUpdate struct {
	Event         string `json:"e"` // Event type
	EventTime     int64  `json:"E"` // Event time in microseconds
	Symbol        string `json:"s"` // Symbol
	ClientOrderID string `json:"c"` // Client order ID
	Side          string `json:"S"` // Side
	OrderType     string `json:"o"` // Order type
	TimeInForce   string `json:"f"` // Time in force
	Quantity      string `json:"q"` // Quantity
	QuoteQuantity string `json:"Q"` // Quantity in quote
	Price         string `json:"p"` // Price
	TriggerPrice  string `json:"P"` // Trigger price
	OrderState    string `json:"X"` // Order state
	OrderID       string `json:"i"` // Order ID
	TradeID       string `json:"t"` // Trade id
	FillQuantity  string `json:"l"` // Fill quantity
	ExecutedQty   string `json:"z"` // Executed quantity
	ExecutedQtyQ  string `json:"Z"` // Executed quantity in quote
	FillPrice     string `json:"L"` // Fill price
	IsMaker       bool   `json:"m"` // Whether the order was maker
	Fee           string `json:"n"` // Fee
	FeeSymbol     string `json:"N"` // Fee symbol
	SelfTradePrev string `json:"V"` // Self trade prevention
	EngineTime    int64  `json:"T"` // Engine timestamp in microseconds
}

// WebSocketClient表示WebSocket客户端
type WebSocketClient struct {
	conn *websocket.Conn
}

// NewWebSocketClient创建新的WebSocketClient实例
func NewWebSocketClient() (*WebSocketClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}

	return &WebSocketClient{conn: conn}, nil
}

// Subscribe订阅一个流
func (client *WebSocketClient) Subscribe(stream string) error {
	data := struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Method: "SUBSCRIBE",
		Params: []string{stream},
	}

	return client.sendJSON(data)
}

// Unsubscribe表示取消订阅某个流
func (client *WebSocketClient) Unsubscribe(stream string) error {
	data := struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Method: "UNSUBSCRIBE",
		Params: []string{stream},
	}

	return client.sendJSON(data)
}

// sendJSON通过WebSocket连接发送JSON消息
func (client *WebSocketClient) sendJSON(data interface{}) error {
	client.conn.SetWriteDeadline(time.Now().Add(readWait))
	return client.conn.WriteJSON(data)
}

// ListenAndServe监听传入的WebSocket消息
func (client *WebSocketClient) ListenAndServe() {
	defer client.conn.Close()

	for {
		select {
		case <-done:
			return
		default:
			_, message, err := client.conn.ReadMessage()
			if err != nil {
				panic(err)
			}

			log.Printf("Received message: %s\n", message)
		}
	}
}

// 生成签名
func generateSignature(
	instruction,
	apikey,
	secret string,
	payload map[string]string,
) map[string]string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	window := "10000"

	data := make(map[string]string)
	data["instruction"] = instruction
	for k, v := range payload {
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

func handleMessage(message []byte, callback func(orderUpdate OrderUpdate)) {
	var orderUpdate OrderUpdate
	err := json.Unmarshal(message, &orderUpdate)
	if err != nil {
		panic(err)
	}

	// Process the order update event
	callback(orderUpdate)
}

// processOrderUpdate处理订单更新事件
func processOrderUpdate(orderUpdate OrderUpdate) {
	// Print or process the order update event as needed
	fmt.Printf("Received order update event: %+v\n", orderUpdate)
}

func main() {
	// 处理中断信号
	signal.Notify(interrupt, os.Interrupt)
	defer close(done)

	// 创建WebSocket客户端
	client, err := NewWebSocketClient()
	if err != nil {
		panic(err)
	}

	// 订阅流
	if err := client.Subscribe("depth.SOL_USDC"); err != nil {
		panic(err)
	}

	// 监听传入的消息
	go client.ListenAndServe()

	// 等待中断信号
	<-interrupt
	log.Println("Received interrupt signal, closing WebSocket connection...")
}
