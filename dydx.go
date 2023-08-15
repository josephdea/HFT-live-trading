package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type DYDXExchange struct {
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

func (t *Trades) initTradesDYDX(arrTrades []TradeUpdate) { //process initial snapshot of trades
	for i := len(arrTrades) - 1; i >= 0; i-- {
		element := arrTrades[i]
		createTime, _ := time.Parse("2006-01-02T15:04:05.000Z", element.CreatedAt)
		px, _ := strconv.ParseFloat(element.Price, 64)
		side := element.Side
		sz, _ := strconv.ParseFloat(element.Size, 64)
		t.Update(createTime.Local(), px, side, sz)
	}
}
func (t *Trades) updateTradesDYDX(arrTrades []TradeUpdate) { //process subsequent trade messages
	// if len(arrTrades) > 1 {
	// 	fmt.Println("Big Trade Msg")
	// }
	for i := 0; i < len(arrTrades); i++ {
		element := arrTrades[i]
		px, _ := strconv.ParseFloat(element.Price, 64)
		side := element.Side
		sz, _ := strconv.ParseFloat(element.Size, 64)
		if element.Liquidation {
			sz *= -1
		}
		t.Update(time.Now(), px, side, sz)
	}

}
func (t *Trades) DebugupdateDYDXTrades(curTime time.Time, coin string) {
	mid := 0.0
	if coin == "BTC-USD" {
		mid = 30000
	} else {
		mid = 2000
	}
	for i := 499; i >= 0; i-- {
		//for i := 0; i < 500; i++ {
		px := mid + 1.0 + float64(i/100)
		sz := 0.1
		a := time.Duration(-float64(i)*100) * time.Millisecond
		t.Update(curTime.Add(a), px, "sell", sz)

		px = mid - 2.0 - float64(i/100)
		sz = 0.5
		a = time.Duration(-float64(i)*100) * time.Millisecond
		t.Update(curTime.Add(a), px, "buy", sz)
	}

}
func (exch *DYDXExchange) Debug() {
	for _, coin := range exch.SymbolList {
		mid := 0.0
		if coin == "BTC-USD" {
			mid = 30000
		} else {
			mid = 2000
		}
		offset := 0
		ask := make([]PriceLevel, 0)
		bid := make([]PriceLevel, 0)
		for i := 10; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid+float64(i))
			sz_str := fmt.Sprintf("%f", 0.1)
			off_str := strconv.Itoa(offset)
			ask = append(ask, PriceLevel{Price: px_str, Size: sz_str, Offset: off_str})
		}
		for i := 1; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid-float64(i))
			sz_str := fmt.Sprintf("%f", 0.2)
			off_str := strconv.Itoa(offset)
			bid = append(bid, PriceLevel{Price: px_str, Size: sz_str, Offset: off_str})
		}
		exch.orderbooks[coin].initBookDYDX(ask, bid)
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

		exch.orderbooks[coin].updateBookDYDX(update_ask, update_bid, strconv.Itoa(offset))
		exch.orderbooks[coin].IsCoherent()
		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		offset += 1
		for i := 5; i < 10; i++ { //sell add test
			px_str := fmt.Sprintf("%f", mid+float64(i))
			update_ask = append(update_ask, []string{px_str, "0.3"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.3", Side: "sell"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 8; i < 10; i++ { //buy add test
			px_str := fmt.Sprintf("%f", mid-float64(i))
			update_bid = append(update_bid, []string{px_str, "0.3"})
			// ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.3", Side: "buy"}}
			// exch.orderbooks[coin].updateBookGDAX(ss)
		}
		exch.orderbooks[coin].updateBookDYDX(update_ask, update_bid, strconv.Itoa(offset))
		exch.orderbooks[coin].IsCoherent()
		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		offset += 1

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

		exch.orderbooks[coin].updateBookDYDX(update_ask, update_bid, strconv.Itoa(offset))
		exch.orderbooks[coin].IsCoherent()
		update_ask = make([][]string, 0)
		update_bid = make([][]string, 0)
		exch.trades[coin].DebugupdateDYDXTrades(time.Now(), coin)
	}

}

func (ob *Orderbook) initBookDYDX(ask []PriceLevel, bid []PriceLevel) { //initialize initial snapshot message of book
	ob.update(-1.0, -1.0, -1.0)
	ob.offsets = make(map[float64]uint64, 0)
	askVarSafe := make([]Level, 1, len(ask)+100)
	bidVarSafe := make([]Level, 1, len(bid)+100)
	askVarSafe[0] = Level{px: math.MaxFloat64, sz: 1}
	for i := len(ask) - 1; i >= 0; i-- {
		element := ask[i]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)
		offset, _ := strconv.Atoi(element.Offset)
		ob.offsets[price] = uint64(offset)
		if size == 0 {
			continue
		}
		askVarSafe = append(askVarSafe, Level{px: price, sz: size})
		ob.update(price, size, 1.0)
	}
	askVarSafe = append(askVarSafe, Level{px: -1, sz: 1})
	bidVarSafe[0] = Level{px: -1, sz: 1}
	for i := len(bid) - 1; i >= 0; i-- {
		element := bid[i]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)
		offset, _ := strconv.Atoi(element.Offset)

		ob.offsets[price] = uint64(offset)
		if size == 0 {
			continue
		}
		bidVarSafe = append(bidVarSafe, Level{px: price, sz: size})
		ob.update(price, size, -1.0)
	}
	bidVarSafe = append(bidVarSafe, Level{px: math.MaxFloat64, sz: 1})
	ob.ask = askVarSafe
	ob.bid = bidVarSafe
	fmt.Println("orderbook initialized")
	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}
	ob.update(-1.0, -1.0, -1.0)
}

func (ob *Orderbook) updateBookDYDX(ask [][]string, bid [][]string, offset string) { //update book with subsequent book messages
	var lowest_ask_px, highest_bid_px float64
	ask_ptr := len(ob.ask) - 1 //ask_ptr start at end - len(ob.ask-2) is the cushion necessary (cushion necessary have to check insert but can start one left)
	offsetVal, _ := strconv.Atoi(offset)
	offsetUINT := uint64(offsetVal)
	for i := 0; i < len(ask); i++ {
		price, _ := strconv.ParseFloat(ask[i][0], 64)
		size, _ := strconv.ParseFloat(ask[i][1], 64)
		val, ok := ob.offsets[price]
		if ok && val >= offsetUINT {
			continue
		} else {
			ob.offsets[price] = offsetUINT
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
					if ob.CancelFlag {
						ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
							Side:          1,
							Price:         price,
							RemoveSize:    ob.ask[ask_ptr].sz - size,
							RemainingSize: size,
							Level:         len(ob.ask) - 1 - ask_ptr})
					}

				} else if ob.ask[ask_ptr].sz < size && len(ob.ask)-1-ask_ptr <= 20 {
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
				if ob.CancelFlag {
					ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
						Side:          1,
						Price:         price,
						RemoveSize:    ob.ask[ask_ptr].sz,
						RemainingSize: size,
						Level:         len(ob.ask) - 2 - ask_ptr})

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
	for i := 0; i < len(bid); i++ {
		price, _ := strconv.ParseFloat(bid[i][0], 64)
		size, _ := strconv.ParseFloat(bid[i][1], 64)
		val, ok := ob.offsets[price]
		if ok && val >= offsetUINT {
			continue
		} else {
			ob.offsets[price] = offsetUINT
		}
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
				if ob.bid[bid_ptr].sz > size {
					if ob.CancelFlag {
						ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
							Side:          -1,
							Price:         price,
							RemoveSize:    ob.bid[bid_ptr].sz - size,
							RemainingSize: size,
							Level:         len(ob.bid) - 1 - bid_ptr})
					}

				} else if ob.bid[bid_ptr].sz < size && len(ob.bid)-1-bid_ptr <= 20 {
					if ob.AddFlag {
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
				if ob.CancelFlag {
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
		// questionable
	}
	if len(ask) >= 1 {
		lowest_ask_px, _ = strconv.ParseFloat(ask[0][0], 64)
		for ob.bid[len(ob.bid)-2].px >= lowest_ask_px {
			ob.bid = Remove(ob.bid, len(ob.bid)-2)
			//fmt.Println("CROSSED")
		}
	}
	if len(bid) >= 1 {
		highest_bid_px, _ = strconv.ParseFloat(bid[0][0], 64)
		for ob.ask[len(ob.ask)-2].px <= highest_bid_px {
			ob.ask = Remove(ob.ask, len(ob.ask)-2)
			//fmt.Println("CROSSED")
		}
	}
	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}
	//fmt.Println("DYDX mid", ob.Mid())
}

func NewDYDXExchange() *DYDXExchange {
	exch := DYDXExchange{}
	exch.name = "DYDX"
	exch.SymbolList = make([]string, 0)
	exch.orderbooks = make(map[string]*Orderbook, 10)
	exch.trades = make(map[string]*Trades, 10)
	exch.mkts = make(map[string]*Market, 10)
	exch.ctrs = make(map[string]*uint64, 5)
	//exch.ctr = new(uint64)
	return &exch
}

func (exch *DYDXExchange) InitCoin(coin string) {
	exch.SymbolList = append(exch.SymbolList, coin)
	exch.orderbooks[coin] = NewOrderbook()
	exch.trades[coin] = NewTrades()
	exch.mkts[coin] = NewMarket()
	exch.ctrs[coin] = new(uint64)
	*exch.ctrs[coin] = 0
	//exch.coinTrades[coin] = &Trades{coin: coin, limit: tradeLimit, wg: exch.wg, trackAll: track, numMessages: 0}
}

// infinite for loop to listen to websocket msg
// needs to ping if message read somehow (hard?) use a pointer to integer and increment
// if error detected in the pipe, close the original subroutine and launch a new one
func (exch *DYDXExchange) ReadBookMsg(done chan bool, c *websocket.Conn, coin string) {
	init_flag := false
	//var rawData map[string]interface{}
	var message []byte
	var err error
	var UpdateMSG OrderbookUpdate
	var SnapshotMSG OrderbookSnapshot
	var prevID int = -1
	for {
		_, message, err = c.ReadMessage()
		//fmt.Println(string(message))

		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		if init_flag == false { //process first message sent
			err = gojson.Unmarshal(message, &SnapshotMSG)
			if err == nil {
				if SnapshotMSG.MessageID != (prevID + 1) {
					fmt.Println("Error: Missed Message")
				}
				prevID = SnapshotMSG.MessageID
				if SnapshotMSG.Type == "subscribed" {
					//fmt.Println(string(message))
					//fmt.Println(SnapshotMSG.Contents.Bids)
					exch.orderbooks[coin].initBookDYDX(SnapshotMSG.Contents.Asks, SnapshotMSG.Contents.Bids)
					init_flag = true
					// if coin == "BTC-USD" {
					// 	exch.byteArray = append(exch.byteArray, []byte("STARTING")...)
					// 	//xch.orderbooks[coin].Print()
					// 	exch.byteArray = append(exch.byteArray, message...)
					// }

				}
				exch.orderbooks[coin].messageID = SnapshotMSG.MessageID
			}
		} else { //6.2 microseconds
			//start := time.Now()

			gojson.Unmarshal(message, &UpdateMSG)

			if UpdateMSG.MessageID != (prevID + 1) {
				fmt.Println("Error: Missed Message")
			}
			prevID = UpdateMSG.MessageID

			exch.orderbooks[coin].updateBookDYDX(UpdateMSG.Contents.Asks, UpdateMSG.Contents.Bids, UpdateMSG.Contents.Offset)
			// if coin == "BTC-USD" {
			// 	a := []byte(time.Now().String())
			// 	exch.byteArray = append(exch.byteArray, a...)
			// 	exch.byteArray = append(exch.byteArray, message...)
			// 	exch.orderbooks[coin].Print()
			// }
			atomic.AddUint64(exch.ctrs[coin], uint64(len(UpdateMSG.Contents.Asks)+len(UpdateMSG.Contents.Bids)))
			//exch.orderbooks[coin].Print()
			// elapsed := time.Since(start)
			// fmt.Println(elapsed)
			// exch.orderbooks[coin].Print()
			// if exch.orderbooks[coin].IsCoherent() == false {
			// 	os.Exit(1)
			// }

		}

		//fmt.Println(len(exch.orderbooks[coin].ask), len(exch.orderbooks[coin].bid))

	}
}

// infinite for loop to listen to websocket msg
//if error detected in the pipe, close the original subroutine and launch a new one

func (exch *DYDXExchange) ReadTradeMsg(done chan bool, c *websocket.Conn, coin string) {
	init_flag := false
	var SnapShotMSG TradeSnapshotJSON
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		err = gojson.Unmarshal(message, &SnapShotMSG)
		//fmt.Println(string(message))
		if err == nil && SnapShotMSG.MessageID != 0 {
			if init_flag == false {
				exch.trades[coin].initTradesDYDX(SnapShotMSG.Contents.Trades)
				init_flag = true
			} else {
				exch.trades[coin].updateTradesDYDX(SnapShotMSG.Contents.Trades)
				//fmt.Println(exch.orderbooks[coin].Mid())
				// if len(SnapShotMSG.Contents.Trades) >= 2 {
				// 	fmt.Println(string(message))
				// }
				// if math.Abs(exch.trades[coin].updates[len(exch.trades[coin].updates)-1][0]-exch.orderbooks[coin].Mid()) >= 7 {
				// 	fmt.Println("price mismatch", exch.trades[coin].updates[len(exch.trades[coin].updates)-1], exch.orderbooks[coin].Mid())
				// 	os.Exit(1)
				// }
				atomic.AddUint64(exch.ctrs[coin], uint64(len(SnapShotMSG.Contents.Trades)))
			}
			//fmt.Println(string(message))

		}
	}
}

func (exch *DYDXExchange) ReadMarketMsg(done chan bool, c *websocket.Conn) {
	init_ctr := 0
	var SnapShotMSG MarketSnapshotJSON
	var SnapShotMSGUpdate MarketSnapshotJSONUpdate
	for {

		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}
		//fmt.Println(string(message))
		if init_ctr == 0 {
			init_ctr++
		} else if init_ctr == 1 {
			init_ctr++
			err = gojson.Unmarshal(message, &SnapShotMSG)
			if err == nil {
				for i := range exch.SymbolList {
					coin := exch.SymbolList[i]
					info := SnapShotMSG.Contents.Markets[coin]
					exch.mkts[coin].Update(time.Now(), info.IndexPrice, info.OraclePrice, info.NextFundingRate, info.OpenInterest)
				}
			}
		} else {
			err = gojson.Unmarshal(message, &SnapShotMSGUpdate)
			if err == nil {
				for i := range exch.SymbolList {
					coin := exch.SymbolList[i]
					info := SnapShotMSGUpdate.Contents[coin]
					exch.mkts[coin].Update(time.Now(), info.IndexPrice, info.OraclePrice, info.NextFundingRate, info.OpenInterest)
				}
			}
		}

	}
}

func (exch *DYDXExchange) ConnectMarket() {
	for {
		log.Println("dydx websocket connect market")
		url := "wss://api.dydx.exchange/v3/ws"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		//defer c.Close()
		done := make(chan bool)
		go exch.ReadMarketMsg(done, c)                      //concurrent running of function to process messages sent to us
		Req := dydxMarketRequest{"subscribe", "v3_markets"} //, "True"}
		err = c.WriteJSON(Req)                              //subscribe to the channel
		ticker := time.NewTicker(3 * time.Second)           //ping every 10 seconds to maintain connection
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("dydx market reader failure")
				break L
			case t := <-ticker.C:
				exch.startTime = time.Now()
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				if err != nil {
					fmt.Println("dydx market orderbook error")
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

// subscribes to the orderbook websocket
func (exch *DYDXExchange) ConnectOrderbook(coin string) {
	for {
		log.Println("dydx websocket connect orderbook", coin)
		url := "wss://api.dydx.exchange/v3/ws"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		//defer c.Close()
		done := make(chan bool)
		go exch.ReadBookMsg(done, c, coin)                           //concurrent running of function to process messages sent to us
		Req := dydxRequest2{"subscribe", "v3_orderbook", coin, true} //, "True"}
		err = c.WriteJSON(Req)                                       //subscribe to the channel
		ticker := time.NewTicker(3 * time.Second)                    //ping every 10 seconds to maintain connection
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("dydx orderbook reader failure")
				break L
			case t := <-ticker.C:
				exch.startTime = time.Now()
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				if err != nil {
					fmt.Println("dydx websocket orderbook error")
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

// subscribes to the trades websocket
func (exch *DYDXExchange) ConnectTrades(coin string) {
	for {
		log.Println("dydx websocket connect trades", coin)
		url := "wss://api.dydx.exchange/v3/ws"
		interrupt := make(chan os.Signal, 1)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()
		done := make(chan bool)
		//ob := orderbook{listen:true}
		go exch.ReadTradeMsg(done, c, coin)

		Req := dydxRequest{"subscribe", "v3_trades", coin} //, "True"}
		//Req := request2{"subscribe"}
		err = c.WriteJSON(Req)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
	L:
		for {
			select {
			case <-done:
				log.Println("dydx trade reader failure")
				break L
			case t := <-ticker.C:
				//fmt.Println(t.String())
				//err := c.WriteJSON(Req)
				err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
				//err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					fmt.Println("dydx websocket trade error")
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

// calls connect orderbook/connect trades
func (exch *DYDXExchange) ConnectFeed() {
	go exch.ConnectMarket()
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		go exch.ConnectOrderbook(coin)
		go exch.ConnectTrades(coin)

	}
}

func (exch *DYDXExchange) SetRecord(mode bool) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exch.orderbooks[coin].record = mode
		exch.trades[coin].record = mode
	}
}
func (exch *DYDXExchange) DumpRecord(dir string) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exchPath := CarvePathDate(dir, coin, exch.name)
		exch.orderbooks[coin].WriteCSV(exchPath)
		exch.trades[coin].WriteCSV(exchPath)
		exch.mkts[coin].WriteCSV(exchPath)
	}

}
func (exch *DYDXExchange) DumpLengths() {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		fmt.Println(len(exch.orderbooks[coin].updates))
		//fmt.Println(len(exch.trades[coin].updates))
	}
}

type NestedUpdate struct {
	Asks   [][]string `json:"asks"`
	Bids   [][]string `json:"bids"`
	Offset string     `json:"offset"`
}
type OrderbookUpdate struct {
	Contents  NestedUpdate `json:"contents"`
	MessageID int          `json:"message_id"`
}
type PriceLevel struct {
	Size   string `json:"size"`
	Price  string `json:"price"`
	Offset string `json:"offset"`
}
type OrderbookJSON struct {
	Asks []PriceLevel `json:"asks"`
	Bids []PriceLevel `json:"bids"`
}
type OrderbookSnapshot struct {
	Type      string        `json:"type"`
	Contents  OrderbookJSON `json:"contents"`
	MessageID int           `json:"message_id"`
}
type TradeUpdate struct {
	Side        string `json:"side"`
	Size        string `json:"size"`
	Price       string `json:"price"`
	CreatedAt   string `json:"createdAt"`
	Liquidation bool   `json:"liquidation"`
}
type NestedTrades struct {
	Trades []TradeUpdate `json:"trades"`
}
type TradeSnapshotJSON struct {
	Type      string       `json:"type"`
	Contents  NestedTrades `json:"contents"`
	MessageID int          `json:"message_id"`
}
type MarketStruct struct {
	Market          string `json:"market"`
	IndexPrice      string `json:"indexPrice"`
	OraclePrice     string `json:"oraclePrice"`
	NextFundingRate string `json:"nextFundingRate"`
	OpenInterest    string `json:"openInterest"`
}
type MarketContents struct {
	Markets map[string]MarketStruct `json:"markets"`
}

type MarketSnapshotJSONUpdate struct {
	Contents map[string]MarketStruct `json:"contents"`
}
type MarketSnapshotJSON struct {
	Type      string         `json:"type"`
	Contents  MarketContents `json:"contents"`
	MessageID int            `json:"message_id"`
}
type dydxRequest struct {
	Method  string `json:"type"`
	Channel string `json:"channel"`
	Params  string `json:"id"`
}
type dydxRequest2 struct {
	Method         string `json:"type"`
	Channel        string `json:"channel"`
	Params         string `json:"id"`
	IncludeOffsets bool   `json:"includeOffsets"`
}
type dydxMarketRequest struct {
	Method  string `json:"type"`
	Channel string `json:"channel"`
	//Params  string `json:"id"`
}
