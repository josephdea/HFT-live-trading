package main

import (
	//"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	jsonparser "github.com/buger/jsonparser"
	gojson "github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type bookSubscription struct {
	Depth int    `json:"depth"`
	Name  string `json:"name"`
}
type krakObSnapshot struct {
	Asks [][]string `json:"as"`
	Bids [][]string `json:"bs"`
}
type krakObUpdate struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}
type krakObRequest struct {
	Method string           `json:"event"`
	Pair   []string         `json:"pair"`
	Sub    bookSubscription `json:"subscription"`
}

type krakTradeUpdate struct {
	Trades [][]string
}

type karakTradeRequest struct {
	Method string            `json:"event"`
	Pair   []string          `json:"pair"`
	Sub    tradeSubscription `json:"subscription"`
}

type tradeSubscription struct {
	Name string `json:"name"`
}

type KRAKENExchange struct {
	startTime    time.Time
	responseTime time.Time
	name         string
	SymbolList   []string
	orderbooks   map[string]*Orderbook
	trades       map[string]*Trades
	ctrs         map[string]*uint64
}

func NewKRAKENExchange() *KRAKENExchange {
	exch := KRAKENExchange{}
	exch.name = "Kraken"
	exch.SymbolList = make([]string, 0)
	exch.orderbooks = make(map[string]*Orderbook, 10)
	exch.trades = make(map[string]*Trades, 10)
	exch.ctrs = make(map[string]*uint64, 5)
	//exch.ctr = new(uint64)
	return &exch
}

func (exch *KRAKENExchange) InitCoin(coin string) {
	exch.SymbolList = append(exch.SymbolList, coin)
	exch.orderbooks[coin] = NewOrderbook()
	exch.trades[coin] = NewTrades()
	exch.ctrs[coin] = new(uint64)
	*exch.ctrs[coin] = 0
}

func (ob *Orderbook) initBookKRAKEN(ask [][]string, bid [][]string) {
	ob.update(-1.0, -1.0, -1.0)
	ptr := 0
	ask_len := 0
	bid_len := 0
	for ptr < len(ask) {
		sz, _ := strconv.ParseFloat(ask[ptr][1], 64)
		if sz != 0 {
			ask_len += 1
		}
		ptr += 1
	}
	ptr = 0
	for ptr < len(bid) {
		sz, _ := strconv.ParseFloat(bid[ptr][1], 64)
		if sz != 0 {
			bid_len += 1
		}
		ptr += 1
	}
	askVarSafe := make([]Level, ask_len+2, ask_len+100)
	bidVarSafe := make([]Level, bid_len+2, bid_len+100)
	// ob.ask = make([]Level, ask_len+2, 600)
	// ob.bid = make([]Level, bid_len+2, 600)
	askVarSafe[len(askVarSafe)-1] = Level{px: -1, sz: 1}
	//curTime := time.Now()
	ptr = -1
	for i := 0; i < len(ask); i++ {
		element := ask[i]
		price, _ := strconv.ParseFloat(element[0], 64)
		size, _ := strconv.ParseFloat(element[1], 64)

		if size == 0 {
			continue
		} else {
			ptr += 1
		}
		askVarSafe[len(askVarSafe)-2-ptr] = Level{px: price, sz: size}
		// if ob.record {
		// 	ob.updates = append(ob.updates, [3]float64{1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, curTime)
		// }
		ob.update(price, size, 1.0)
	}
	askVarSafe[0] = Level{px: math.MaxFloat64, sz: 1} //cushion the array with extreme values

	bidVarSafe[len(bidVarSafe)-1] = Level{px: math.MaxFloat64, sz: 1}
	ptr = -1
	for i := 0; i < len(bid); i++ {
		element := bid[i]
		price, _ := strconv.ParseFloat(element[0], 64)
		size, _ := strconv.ParseFloat(element[1], 64)

		if size == 0 {
			continue
		} else {
			ptr += 1
		}
		bidVarSafe[len(bidVarSafe)-2-ptr] = Level{px: price, sz: size}
		// if ob.record {
		// 	ob.updates = append(ob.updates, [3]float64{-1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, curTime)
		// }
		ob.update(price, size, -1.0)
	}

	bidVarSafe[0] = Level{px: -1, sz: 1} //cushion the array with extreme values
	ob.ask = askVarSafe
	ob.bid = bidVarSafe
	ob.update(-1.0, -1.0, -1.0)
}
func (ob *Orderbook) updateBookKRAKEN(ask [][3]float64, bid [][3]float64) {
	var lowest_ask_px, highest_bid_px float64
	var ask_ptr, bid_ptr int
	//curTime := time.Now()
	if ask == nil {
		goto SKIPASKS
	}
	ask_ptr = len(ob.ask) - 1 //ask_ptr start at end - len(ob.ask-2) is the cushion necessary (cushion necessary have to check insert but can start one left)
	//curTime := time.Now()
	for i := 0; i < len(ask); i++ {
		price := ask[i][0]
		size := ask[i][1]
		var found bool = false
		for ; ; ask_ptr-- {
			curPrice := ob.ask[ask_ptr].px
			if curPrice == price {
				found = true
				break
			} else if price < curPrice {
				break
			}
		}
		if found {
			if size != 0 {
				ob.ask[ask_ptr].sz = size
			} else {
				ob.ask = Remove(ob.ask, ask_ptr)
			}
		} else {
			if size != 0 {
				ob.ask = Insert(ob.ask, ask_ptr+1, Level{px: price, sz: size})
				ask_ptr += 1
			}
		}
		// if ob.record {
		// 	//s := fmt.Sprintf("%f", 123.456)
		// 	ob.updates = append(ob.updates, [3]float64{1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, time.Now())
		// }
		//fmt.Println("debug")
		ob.update(price, size, 1.0)

	}

SKIPASKS:
	if bid == nil {
		goto SKIPBIDS
	}
	bid_ptr = len(ob.bid) - 1
	for i := 0; i < len(bid); i++ {

		price := bid[i][0]
		size := bid[i][1]
		var found bool = false

		for ; ; bid_ptr-- {
			curPrice := ob.bid[bid_ptr].px //error here ob.bid[0] = Level{px: -1, sz: 1}
			if curPrice == price {
				found = true
				break
			} else if price > curPrice {
				break
			}
		}
		if found {
			if size != 0 {
				ob.bid[bid_ptr].sz = size
			} else {
				ob.bid = Remove(ob.bid, bid_ptr)
			}
		} else {
			if size != 0 {
				ob.bid = Insert(ob.bid, bid_ptr+1, Level{px: price, sz: size})
				bid_ptr += 1
			}
		}
		// if ob.record {
		// 	//s := fmt.Sprintf("%f", 123.456)
		// 	ob.updates = append(ob.updates, [3]float64{-1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, time.Now())
		// }
		ob.update(price, size, -1.0)
		// questionable
	}
SKIPBIDS:
	if len(ask) >= 1 {
		lowest_ask_px = ask[0][0]
		for ob.bid[len(ob.bid)-2].px >= lowest_ask_px {
			ob.bid = Remove(ob.bid, len(ob.bid)-2)
			//fmt.Println("CROSSED")
		}
	}
	if len(bid) >= 1 {
		highest_bid_px = bid[0][0]
		for ob.ask[len(ob.ask)-2].px <= highest_bid_px {
			ob.ask = Remove(ob.ask, len(ob.ask)-2)
			//fmt.Println("CROSSED")
		}
	}
}

func (exch *KRAKENExchange) ReadBookMsg(done chan bool, c *websocket.Conn) {
	defer close(done)
	//var rawData []interface{}
	var message []byte
	var err error
	//initflag := false
	var krakUpdateMSG krakObUpdate
	var karakSnapshotMSG krakObSnapshot
	var coin string
	var initctr int = 0
	//var avg []float64
	for {
		_, message, err = c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		//fmt.Println(string(message))
		if (initctr == len(exch.SymbolList) || len(message) < 800) && string(message[0]) == "[" {
			//start := time.Now()
			krakUpdateMSG = krakObUpdate{}
			arrayCtr := 0
			jsonparser.ArrayEach(message, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				if arrayCtr == 1 {
					gojson.Unmarshal(value, &krakUpdateMSG)
				} else {
					coin = string(value)
				}
				arrayCtr += 1
			})

			askFloats := make([][3]float64, 0, 5)
			for i := range krakUpdateMSG.Asks {
				element := krakUpdateMSG.Asks[i]
				px, _ := strconv.ParseFloat(element[0], 64)
				sz, _ := strconv.ParseFloat(element[1], 64)
				t, _ := strconv.ParseFloat(element[2], 64)
				askFloats = append(askFloats, [3]float64{px, sz, t})
			}
			bidFloats := make([][3]float64, 0, 5)
			for i := range krakUpdateMSG.Bids {
				element := krakUpdateMSG.Bids[i]
				px, _ := strconv.ParseFloat(element[0], 64)
				sz, _ := strconv.ParseFloat(element[1], 64)
				t, _ := strconv.ParseFloat(element[2], 64)
				bidFloats = append(bidFloats, [3]float64{px, sz, t})
			}
			sort.Slice(askFloats, func(i, j int) bool {
				i_float := askFloats[i][0]
				j_float := askFloats[j][0]
				if i_float == j_float {
					a := askFloats[i][2]
					b := askFloats[j][2]
					return a < b
				}
				return i_float < j_float
			})
			sort.Slice(bidFloats, func(i, j int) bool {
				i_float := bidFloats[i][0]
				j_float := bidFloats[j][0]
				if i_float == j_float {
					a := bidFloats[i][2]
					b := bidFloats[j][2]
					return a < b
				}
				return i_float > j_float
			})
			exch.orderbooks[coin].updateBookKRAKEN(askFloats, bidFloats)
			//exch.orderbooks[coin].Print()
		} else {
			if string(message[0]) == "[" {
				//fmt.Println(string(message))
				arrayCtr := 0
				jsonparser.ArrayEach(message, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					if arrayCtr == 1 {
						gojson.Unmarshal(value, &karakSnapshotMSG)
					} else {
						coin = string(value)
					}
					arrayCtr += 1
				})
				fmt.Println(coin, "KRAKEN book initialized")
				initctr += 1
				exch.orderbooks[coin].initBookKRAKEN(karakSnapshotMSG.Asks, karakSnapshotMSG.Bids)

			}
		}

	}
}

func (exch *KRAKENExchange) ConnectOrderbook() {
	//exch.initFlag = false
	for {
		log.Println("kraken websocket connect orderbook")
		url := "wss://ws.kraken.com"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		//defer c.Close()
		if err != nil {
			log.Fatal("dial:", err)
			continue
		}

		done := make(chan bool)
		go exch.ReadBookMsg(done, c)                                                                   //concurrent running of function to process messages sent to us
		Req := krakObRequest{"subscribe", exch.SymbolList, bookSubscription{Depth: 100, Name: "book"}} //, "True"}
		err = c.WriteJSON(Req)                                                                         //subscribe to the channel
		ticker := time.NewTicker(10 * time.Second)                                                     //ping every 10 seconds to maintain connection
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("kraken orderbook reader failure")
				break L
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				if err != nil {
					log.Println("kraken websocket orderbook error")
					log.Println("write:", err)
					//done <- true
					break L
				}
			case <-interrupt:
				log.Println("interrupt")
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					break L
				}
			}
		}
		c.Close()
		time.Sleep(time.Millisecond * 200)
	}
}

func (exch *KRAKENExchange) ReadTradeMsg(done chan bool, c *websocket.Conn) {
	defer close(done)
	var rawData []interface{}
	var message []byte
	var err error
	for {
		_, message, err = c.ReadMessage()
		rawData = []interface{}{}
		//fmt.Println("new message", string(message))
		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		if err := gojson.Unmarshal(message, &rawData); err != nil {
			//fmt.Println(err)
			//fmt.Println(string(message))
			//os.Exit(0)
		}
		//fmt.Println("HERE")
		//fmt.Println("trade message", string(message))
		if err == nil && len(rawData) == 4 && rawData[2].(string) == "trade" {
			coin := rawData[3].(string)

			tradesData := rawData[1].([]interface{})

			tradeFloat := make([][4]float64, 0, 5)
			for i := range tradesData {
				element := tradesData[i].([]interface{})
				px, _ := strconv.ParseFloat(element[0].(string), 64)
				sz, _ := strconv.ParseFloat(element[1].(string), 64)
				time, _ := strconv.ParseFloat(element[2].(string), 64)
				side := element[3].(string)
				if side == "s" {
					tradeFloat = append(tradeFloat, [4]float64{time, px, sz, 1.0})
				} else {
					tradeFloat = append(tradeFloat, [4]float64{time, px, sz, -1.0})
				}
			}
			sort.Slice(tradeFloat, func(i, j int) bool {
				i_float := tradeFloat[i][0]
				j_float := tradeFloat[j][0]
				return i_float < j_float
			})
			for i := range tradeFloat {
				element := tradeFloat[i]
				if element[3] == 1.0 {
					exch.trades[coin].Update(time.Now(), element[1], "sell", element[2])
				} else {
					exch.trades[coin].Update(time.Now(), element[1], "buy", element[2])
				}

			}

		}
		//fmt.Println("new message", string(message))

	}
}

func (exch *KRAKENExchange) ConnectTrades() {
	for {
		log.Println("kraken websocket connect trades")
		url := "wss://ws.kraken.com"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		//defer c.Close()
		if err != nil {
			log.Fatal("dial:", err)
			continue
		}

		done := make(chan bool)
		go exch.ReadTradeMsg(done, c)                                                            //concurrent running of function to process messages sent to us
		Req := karakTradeRequest{"subscribe", exch.SymbolList, tradeSubscription{Name: "trade"}} //, "True"}
		err = c.WriteJSON(Req)                                                                   //subscribe to the channel
		ticker := time.NewTicker(10 * time.Second)                                               //ping every 10 seconds to maintain connection
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("kraken trade reader failure")
				break L
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				if err != nil {
					log.Println("kraken websocket trade error")
					log.Println("write:", err)
					break L
				}
			case <-interrupt:
				log.Println("interrupt")
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					break L
				}
				// select {
				// case <-done:
				// case <-time.After(time.Second):
				// }
				// break
			}
		}
		c.Close()
		time.Sleep(time.Millisecond * 200)
	}
}

func (exch *KRAKENExchange) ConnectFeed() {

	go exch.ConnectOrderbook()
	go exch.ConnectTrades()

}

func (exch *KRAKENExchange) SetRecord(mode bool) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exch.orderbooks[coin].record = mode
		exch.trades[coin].record = mode
	}
}

func (exch *KRAKENExchange) DumpRecord(dir string) {
	//permissions := 0644 // or whatever you need
	// err := os.WriteFile("debug.txt", exch.byteArray, 0666)
	// if err != nil {
	// 	// handle error
	// }
	// os.Exit(1)
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		var exchPath string
		if coin[0:3] == "XBT" {
			exchPath = CarvePathDate(dir, "BTC-USD", exch.name)
		} else {
			exchPath = CarvePathDate(dir, strings.Replace(coin, "/", "-", 3), exch.name)
		}

		exch.orderbooks[coin].WriteCSV(exchPath)
		exch.trades[coin].WriteCSV(exchPath)
	}

}
func (exch *KRAKENExchange) DumpLengths() {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		//fmt.Println(len(exch.orderbooks[coin].updateTimes))
		fmt.Println(len(exch.orderbooks[coin].updates))
	}
}
