package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type BNBExchange struct {
	startTime    time.Time
	responseTime time.Time
	name         string
	SymbolList   []string
	orderbooks   map[string]*Orderbook
	trades       map[string]*Trades
	ctrs         map[string]*uint64
	mkts         map[string]*Market
	byteArray    []byte
}

type BNBOrderbookSnapshot struct {
	LastUpdateId int        `json:"lastUpdateId"`
	Asks         [][]string `json:"asks"`
	Bids         [][]string `json:"bids"`
}
type BNBOrderbookContentsUpdate struct {
	Ucaps int        `json:"U"`
	Ulow  int        `json:"u"`
	PU    int        `json:"pu"`
	Asks  [][]string `json:"a"`
	Bids  [][]string `json:"b"`
}
type BNBOrderbookUpdate struct {
	Data BNBOrderbookContentsUpdate `json:"data"`
}
type BNBTradeData struct {
	Side      bool   `json:"m"`
	Size      string `json:"q"`
	Price     string `json:"p"`
	CreatedAt string `json:"createdAt"`
}
type BNBTradeUpdate struct {
	Data BNBTradeData `json:"data"`
}

func NewBNBExchange() *BNBExchange {
	exch := BNBExchange{}
	exch.name = "BNB"
	exch.SymbolList = make([]string, 0)
	exch.orderbooks = make(map[string]*Orderbook, 10)
	exch.trades = make(map[string]*Trades, 10)
	exch.mkts = make(map[string]*Market, 10)
	exch.ctrs = make(map[string]*uint64, 5)
	return &exch
}
func (exch *BNBExchange) InitCoin(coin string) {
	exch.SymbolList = append(exch.SymbolList, coin)
	exch.orderbooks[coin] = NewOrderbook()
	exch.trades[coin] = NewTrades()
	exch.mkts[coin] = NewMarket()
	exch.ctrs[coin] = new(uint64)
	*exch.ctrs[coin] = 0
	//exch.coinTrades[coin] = &Trades{coin: coin, limit: tradeLimit, wg: exch.wg, trackAll: track, numMessages: 0}
}

func (ob *Orderbook) initBookBNB(asks [][]string, bids [][]string, messageid int) {
	ob.messageID = messageid
	ob.update(-1.0, -1.0, -1.0)
	ob.offsets = make(map[float64]uint64, 0)
	//curTime := time.Now()
	askVarSafe := make([]Level, 1, len(asks)+100)
	bidVarSafe := make([]Level, 1, len(bids)+100)
	// ob.ask = make([]Level, 1, len(ask)+100)
	// ob.bid = make([]Level, 1, len(bid)+100)

	askVarSafe[0] = Level{px: math.MaxFloat64, sz: 1}

	for i := len(asks) - 1; i >= 0; i-- {
		element := asks[i]
		price, _ := strconv.ParseFloat(element[0], 64)
		size, _ := strconv.ParseFloat(element[1], 64)
		if size == 0 {
			continue
		}
		askVarSafe = append(askVarSafe, Level{px: price, sz: size})
		ob.update(price, size, 1.0)
	}
	askVarSafe = append(askVarSafe, Level{px: -1, sz: 1})

	bidVarSafe[0] = Level{px: -1, sz: 1}

	for i := len(bids) - 1; i >= 0; i-- {
		element := bids[i]
		price, _ := strconv.ParseFloat(element[0], 64)
		size, _ := strconv.ParseFloat(element[1], 64)
		if size == 0 {
			continue
		}
		bidVarSafe = append(bidVarSafe, Level{px: price, sz: size})
		ob.update(price, size, -1.0)

	}
	bidVarSafe = append(bidVarSafe, Level{px: math.MaxFloat64, sz: 1})

	ob.ask = askVarSafe
	ob.bid = bidVarSafe
	fmt.Println("BNB orderbook initialized")
	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}
	ob.update(-1.0, -1.0, -1.0)
	// if ob.IsCoherent() == false {
	// 	fmt.Println("INIT NOT COHERENT")
	// 	os.Exit(1)
	// }

}

func (exch *BNBExchange) Debug() {
	for _, coin := range exch.SymbolList {
		mid := 0.0
		if coin == "BTC-USD" {
			mid = 30000
		} else {
			mid = 2000
		}
		offset := 0
		ask := make([][]string, 0)
		bid := make([][]string, 0)
		// price, _ := strconv.ParseFloat(ask[i][0], 64)
		// size, _ := strconv.ParseFloat(ask[i][1], 64)
		for i := 10; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid+float64(i))
			sz_str := fmt.Sprintf("%f", 0.1)
			ask = append(ask, []string{px_str, sz_str})
		}
		for i := 1; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid-float64(i))
			sz_str := fmt.Sprintf("%f", 0.2)
			bid = append(bid, []string{px_str, sz_str})

		}

		exch.orderbooks[coin].initBookBNB(ask, bid, 0)
		if coin == "BTC-USD" {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}

		exch.orderbooks[coin].IsCoherent()
		offset += 1
		update_ask := make([][]string, 0)
		update_bid := make([][]string, 0)
		for i := 1; i < 10; i++ {
			px_str := fmt.Sprintf("%f", mid+float64(i)) //sell cancel test
			update_ask = append(update_ask, []string{px_str, "0.0"})
		}
		for i := 1; i < 12; i++ { //buy cancel test
			px_str := fmt.Sprintf("%f", mid-float64(i))
			update_bid = append(update_bid, []string{px_str, "0.0"})
		}

		exch.orderbooks[coin].updateBookBNB(update_ask, update_bid, -1, 1, 15)
		if coin == "BTC-USD" {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		exch.orderbooks[coin].IsCoherent()
		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		offset += 1
		for i := 5; i < 25; i++ { //sell add test
			px_str := fmt.Sprintf("%f", mid+float64(i))
			update_ask = append(update_ask, []string{px_str, "0.5"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.3", Side: "sell"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 24; i >= 8; i-- {
			//for i := 8; i < 25; i++ { //buy add test
			px_str := fmt.Sprintf("%f", mid-float64(i))
			update_bid = append(update_bid, []string{px_str, "0.5"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.3", Side: "buy"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}

		exch.orderbooks[coin].updateBookBNB(update_ask, update_bid, -1, 1, 15)
		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		offset += 1
		exch.orderbooks[coin].IsCoherent()
		for i := 4; i >= 2; i-- {
			px_str := fmt.Sprintf("%f", mid+float64(i))
			update_ask = append(update_ask, []string{px_str, "0.1"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.1", Side: "sell"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 4; i >= 2; i-- {
			px_str := fmt.Sprintf("%f", mid-float64(i))
			update_bid = append(update_bid, []string{px_str, "0.1"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.1", Side: "buy"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}

		exch.orderbooks[coin].updateBookBNB(update_ask, update_bid, -1, 1, 15)
		exch.orderbooks[coin].IsCoherent()

		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		exch.trades[coin].DebugupdateDYDXTrades(time.Now(), coin)
	}
	// fmt.Println(exch.orderbooks["BTC-USD"].bid)
	// fmt.Println(exch.orderbooks["BTC-USD"].IsCoherent())
	// os.Exit(1)
	//ask []coinbasepro.SnapshotEntry, bid []coinbasepro.SnapshotEntry
}

func (ob *Orderbook) updateBookBNB(ask [][]string, bid [][]string, U int, u int, pu int) bool {
	if u < ob.messageID {
		return true
	}
	if U <= ob.messageID && u >= ob.messageID {

	} else {
		if pu != ob.pu {
			fmt.Println("Incorrect previous u")
			return false
		}
	}

	ob.pu = u
	var lowest_ask_px, highest_bid_px float64
	lowest_ask_px = math.MaxFloat64
	highest_bid_px = -1.0
	// Drop any event where u is < lastUpdateId in the snapshot.
	// The first processed event should have U <= lastUpdateId AND u >= lastUpdateId
	// While listening to the stream, each new event's pu should be equal to the previous event's u, otherwise initialize the process from step 3.
	ask_ptr := len(ob.ask) - 1 //ask_ptr start at end - len(ob.ask-2) is the cushion necessary (cushion necessary have to check insert but can start one left)
	//curTime := time.Now()
	for i := 0; i < len(ask); i++ {
		//for i := 0; i < len(ask); i++ {
		price, _ := strconv.ParseFloat(ask[i][0], 64)
		size, _ := strconv.ParseFloat(ask[i][1], 64)
		if size != 0 {
			lowest_ask_px = math.Min(lowest_ask_px, price)
		}
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
				if ob.ask[ask_ptr].sz > size {
					if ob.CancelFlag && len(ob.ask)-1-ask_ptr <= 20 {
						//fmt.Println(time.Now(), "BNB Cancel", price, "sell", len(ob.ask)-1-ask_ptr, ob.ask[ask_ptr].sz-size)
						ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
							Side:          1,
							Price:         price,
							RemoveSize:    ob.ask[ask_ptr].sz - size,
							RemainingSize: size,
							Level:         len(ob.ask) - 1 - ask_ptr})
					}

				} else if ob.ask[ask_ptr].sz < size && len(ob.ask)-1-ask_ptr <= 20 {
					// fmt.Println(time.Now(), "BNB Add", price, "sell", len(ob.ask)-1-ask_ptr)
					if ob.AddFlag {
						ob.RecentAdds.PushBack(RecentAdd{timestamp: time.Now(),
							Side:     1,
							Price:    price,
							AddSize:  size - ob.ask[ask_ptr].sz,
							OrigSize: ob.ask[ask_ptr].sz,
							Level:    len(ob.ask) - 1 - ask_ptr})
					}

				}
				ob.ask[ask_ptr].sz = size
			} else {
				if ob.CancelFlag && len(ob.ask)-1-ask_ptr <= 20 {
					//fmt.Println(time.Now(), "BNB Cancel", price, "sell", len(ob.ask)-1-ask_ptr, ob.ask[ask_ptr].sz)
					ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
						Side:          1,
						Price:         price,
						RemoveSize:    ob.ask[ask_ptr].sz,
						RemainingSize: size,
						Level:         len(ob.ask) - 1 - ask_ptr})
				}

				ob.ask = Remove(ob.ask, ask_ptr)
			}
		} else {
			lvl := len(ob.ask) - 2 - ask_ptr
			if lvl != 0 {
				lvl += 1
			}
			if size != 0 && lvl <= 20 {
				if ob.AddFlag {
					// fmt.Println(time.Now(), "BNB Add", price, "sell", lvl)
					ob.RecentAdds.PushBack(RecentAdd{timestamp: time.Now(),
						Side:     1,
						Price:    price,
						AddSize:  size,
						OrigSize: 0,
						Level:    lvl})
				}

				if lvl == 0 && ob.CutinFlag {
					ob.RecentCutIns.PushBack(CutinOrder{timestamp: time.Now(), side: 1, size: size, spread: ob.Spread(), cutin: ob.BestAsk() - price})
				}
				ob.ask = Insert(ob.ask, ask_ptr+1, Level{px: price, sz: size})
				ask_ptr += 1
			}

		}
		ob.update(price, size, 1.0)

	}

	bid_ptr := len(ob.bid) - 1
	for i := len(bid) - 1; i >= 0; i-- {
		//for i := 0; i < len(bid); i++ {
		price, _ := strconv.ParseFloat(bid[i][0], 64)
		size, _ := strconv.ParseFloat(bid[i][1], 64)
		if size != 0 {
			highest_bid_px = math.Max(highest_bid_px, price)
		}
		var found bool = false
		for ; ; bid_ptr-- {
			curPrice := ob.bid[bid_ptr].px //error here ob.bid[0] = Level{px: -1, sz: 1} bidptr is somehow length 45728
			if curPrice == price {
				found = true
				break
			} else if price > curPrice {
				break
			}
		}
		if found {
			if size != 0 {
				if ob.bid[bid_ptr].sz > size {
					if ob.CancelFlag && len(ob.bid)-1-bid_ptr <= 20 {
						//fmt.Println(time.Now(), "BNB Cancel", price, "sell", len(ob.bid)-1-bid_ptr, ob.bid[bid_ptr].sz-size)
						ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
							Side:          -1,
							Price:         price,
							RemoveSize:    ob.bid[bid_ptr].sz - size,
							RemainingSize: size,
							Level:         len(ob.bid) - 1 - bid_ptr})
					}

				} else if ob.bid[bid_ptr].sz < size && len(ob.bid)-1-bid_ptr <= 20 {
					if ob.AddFlag {

						// fmt.Println(time.Now(), "BNB Add", price, "buy", len(ob.bid)-1-bid_ptr)
						ob.RecentAdds.PushBack(RecentAdd{timestamp: time.Now(),
							Side:     -1,
							Price:    price,
							AddSize:  size - ob.bid[bid_ptr].sz,
							OrigSize: ob.bid[bid_ptr].sz,
							Level:    len(ob.bid) - 1 - bid_ptr})
					}

				}
				ob.bid[bid_ptr].sz = size
			} else {
				if ob.CancelFlag && len(ob.bid)-1-bid_ptr <= 20 {
					//fmt.Println(time.Now(), "BNB Cancel", price, "sell", len(ob.bid)-1-bid_ptr, ob.bid[bid_ptr].sz)
					ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
						Side:          -1,
						Price:         price,
						RemoveSize:    ob.bid[bid_ptr].sz,
						RemainingSize: size,
						Level:         len(ob.bid) - 1 - bid_ptr})
				}

				ob.bid = Remove(ob.bid, bid_ptr)
			}
		} else {
			lvl := len(ob.bid) - 2 - bid_ptr
			if lvl != 0 {
				lvl += 1
			}
			if size != 0 && lvl <= 20 {
				if ob.AddFlag {
					// fmt.Println(time.Now(), "BNB Add", price, "buy", lvl)
					ob.RecentAdds.PushBack(RecentAdd{timestamp: time.Now(),
						Side:     -1,
						Price:    price,
						AddSize:  size,
						OrigSize: 0,
						Level:    lvl})
				}

				if lvl == 0 && ob.CutinFlag {
					ob.RecentCutIns.PushBack(CutinOrder{timestamp: time.Now(), side: -1, size: size, spread: ob.Spread(), cutin: price - ob.BestBid()})
				}
				ob.bid = Insert(ob.bid, bid_ptr+1, Level{px: price, sz: size})
				bid_ptr += 1
			}

		}
		ob.update(price, size, -1.0)

	}
	if len(ask) >= 1 {
		for ob.bid[len(ob.bid)-2].px >= lowest_ask_px {
			ob.bid = Remove(ob.bid, len(ob.bid)-2)
		}
	}
	if len(bid) >= 1 {
		for ob.ask[len(ob.ask)-2].px <= highest_bid_px {
			ob.ask = Remove(ob.ask, len(ob.ask)-2)
		}
	}

	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}
	return true
}

func (exch *BNBExchange) ReadBookMsg(done chan bool, c *websocket.Conn, coin string) {

	var message []byte
	var err error
	var UpdateMSG BNBOrderbookUpdate
	var SnapshotMSG BNBOrderbookContentsUpdate
	for {
		_, message, err = c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		err = gojson.Unmarshal(message, &UpdateMSG)
		SnapshotMSG = UpdateMSG.Data
		res := exch.orderbooks[coin].updateBookBNB(SnapshotMSG.Asks, SnapshotMSG.Bids, SnapshotMSG.Ucaps, SnapshotMSG.Ulow, SnapshotMSG.PU)
		if res == false {
			fmt.Println("BNB message order error")
			done <- true
			return
		}

		atomic.AddUint64(exch.ctrs[coin], uint64(len(SnapshotMSG.Asks)+len(SnapshotMSG.Bids)))

	}
}

func (exch *BNBExchange) ConnectOrderbook(coin string) {
	format_str := strings.ToLower(strings.Replace(coin, "-", "", -1))
	for {
		log.Println("bnb websocket connect orderbook", coin)
		//<symbol>@depth@100ms
		baseurl1 := "wss://fstream.binance.com/stream?streams=" + format_str + "t@depth"
		url := baseurl1 //+ format_str + "@depth"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()
		done := make(chan bool)

		req, err := http.Get("https://fapi.binance.com/fapi/v1/depth?symbol=" + format_str + "T&limit=1000") //?market=ETH-USD
		if err != nil {
			fmt.Println("ERROR")
			log.Fatalln(err)
		}
		body, err := ioutil.ReadAll(req.Body)
		var SnapshotMSG BNBOrderbookSnapshot
		err = gojson.Unmarshal(body, &SnapshotMSG)

		exch.orderbooks[coin].initBookBNB(SnapshotMSG.Asks, SnapshotMSG.Bids, SnapshotMSG.LastUpdateId)
		//update book
		//read messages
		go exch.ReadBookMsg(done, c, coin)
		ticker := time.NewTicker(3 * time.Second) //ping every 10 seconds to maintain connection
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("bnb orderbook reader failure")
				break L
			case t := <-ticker.C:
				exch.startTime = time.Now()
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				if err != nil {
					fmt.Println("bnb websocket orderbook error")
					log.Println("write:", err)
					break L
				}
			case <-interrupt:
				log.Println("interrupt")
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				// select {
				// case <-done:
				// case <-time.After(time.Second):
				// }
				return
			}
		}
		c.Close()
		time.Sleep(200 * time.Millisecond)
	}
}

func (exch *BNBExchange) ReadTradeMsg(done chan bool, c *websocket.Conn, coin string) {
	// init_flag := false
	//var err error
	var SnapShotMSG BNBTradeUpdate
	var data BNBTradeData
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		//fmt.Println(string(message))
		err = gojson.Unmarshal(message, &SnapShotMSG)
		//fmt.Println(err)
		if err == nil {
			data = SnapShotMSG.Data
			px, _ := strconv.ParseFloat(data.Price, 64)
			sz, _ := strconv.ParseFloat(data.Size, 64)
			if data.Side == false { //executed on ask side
				exch.trades[coin].Update(time.Now(), px, "ask", sz)
			} else {
				//fmt.Println("bid trade update")
				exch.trades[coin].Update(time.Now(), px, "bid", sz)
			}
			atomic.AddUint64(exch.ctrs[coin], 1)
		}

	}
}

func (exch *BNBExchange) ConnectTrades(coin string) {
	format_str := strings.ToLower(strings.Replace(coin, "-", "", -1))
	for {
		log.Println("bnb websocket connect trades", coin)
		baseurl1 := "wss://fstream.binance.com/stream?streams=" + format_str + "t@aggTrade" //"wss://fstream.binance.com/ws"
		//baseurl2 := "wss://stream.binance.us:9443/ws/"
		url := baseurl1 //+ format_str + "@depth"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()
		done := make(chan bool)
		go exch.ReadTradeMsg(done, c, coin)
		//Req := request2{"subscribe"}
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("bnb trade reader failure")
				break L
			case t := <-ticker.C:
				//fmt.Println(t.String())
				//err := c.WriteJSON(Req)
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				//err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					fmt.Println("bnb websocket trade error")
					log.Println("write:", err)
					break L
				}

			case <-interrupt:
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
			}
		}
		c.Close()
		time.Sleep(200 * time.Millisecond)
	}
}
func (exch *BNBExchange) ConnectFeed() {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		go exch.ConnectOrderbook(coin)
		go exch.ConnectTrades(coin)
	}
}

func (exch *BNBExchange) SetRecord(mode bool) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exch.orderbooks[coin].record = mode
		exch.trades[coin].record = mode
	}
}
func (exch *BNBExchange) DumpRecord(dir string) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exchPath := CarvePathDate(dir, coin, exch.name)
		exch.orderbooks[coin].WriteCSV(exchPath)
		exch.trades[coin].WriteCSV(exchPath)
	}

}
func (exch *BNBExchange) DumpLengths() {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		fmt.Println(len(exch.orderbooks[coin].updates))
		//fmt.Println(len(exch.trades[coin].updates))
	}
}
