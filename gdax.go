package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
)

type GDAXExchange struct {
	startTime    time.Time
	responseTime time.Time
	name         string
	SymbolList   []string
	orderbooks   map[string]*Orderbook
	trades       map[string]*Trades
	ctrs         map[string]*uint64
}

func NewGDAXExchange() *GDAXExchange {
	exch := GDAXExchange{}
	exch.name = "GDAX"
	exch.SymbolList = make([]string, 0)
	exch.orderbooks = make(map[string]*Orderbook)
	exch.trades = make(map[string]*Trades)
	exch.ctrs = make(map[string]*uint64, 5)
	//exch.coinTrades = make(map[string]*Trades)
	return &exch
}

func (exch *GDAXExchange) InitCoin(coin string) {
	exch.SymbolList = append(exch.SymbolList, coin)
	exch.orderbooks[coin] = NewOrderbook()
	exch.trades[coin] = NewTrades()
	exch.ctrs[coin] = new(uint64)
	*exch.ctrs[coin] = 0

}
func (t *Trades) updateGDAXTrades(trade coinbasepro.Message) {

	//createTime := trade.Time.Time()
	px, _ := strconv.ParseFloat(trade.Price, 64)
	side := trade.Side
	sz, _ := strconv.ParseFloat(trade.Size, 64)
	//fmt.Println((createTime.Sub(time.Now())))
	t.Update(time.Now(), px, side, sz)
}
func (t *Trades) DebugupdateGDAXTrades(curTime time.Time, coin string) {
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
		sz = 0.2
		a = time.Duration(-float64(i)*100) * time.Millisecond
		t.Update(curTime.Add(a), px, "buy", sz)
	}

}
func (exch *GDAXExchange) Debug() {
	for _, coin := range exch.SymbolList {
		mid := 0.0
		if coin == "BTC-USD" {
			mid = 30000
		} else {
			mid = 2000
		}
		ask := make([]coinbasepro.SnapshotEntry, 0)
		bid := make([]coinbasepro.SnapshotEntry, 0)
		for i := 1; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid+float64(i))
			sz_str := fmt.Sprintf("%f", 0.1)
			ask = append(ask, coinbasepro.SnapshotEntry{Price: px_str, Size: sz_str})
		}
		for i := 10; i < 100; i++ {
			px_str := fmt.Sprintf("%f", mid-float64(i))
			sz_str := fmt.Sprintf("%f", 0.2)
			bid = append(bid, coinbasepro.SnapshotEntry{Price: px_str, Size: sz_str})
		}
		exch.orderbooks[coin].initBookGDAX(ask, bid)
		for i := 1; i < 10; i++ { //sell cancel test
			px_str := fmt.Sprintf("%f", mid+float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.0", Side: "sell"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 1; i < 12; i++ { //buy cancel test
			px_str := fmt.Sprintf("%f", mid-float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.0", Side: "buy"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 5; i < 15; i++ { //sell add test
			px_str := fmt.Sprintf("%f", mid+float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.5", Side: "sell"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 8; i < 15; i++ { //buy add test
			px_str := fmt.Sprintf("%f", mid-float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.5", Side: "buy"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 4; i >= 2; i-- {
			px_str := fmt.Sprintf("%f", mid+float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.1", Side: "sell"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		for i := 4; i >= 2; i-- {
			px_str := fmt.Sprintf("%f", mid-float64(i))
			ss := []coinbasepro.SnapshotChange{coinbasepro.SnapshotChange{Price: px_str, Size: "0.1", Side: "buy"}}
			exch.orderbooks[coin].updateBookGDAX(ss)
		}
		exch.trades[coin].DebugupdateGDAXTrades(time.Now(), coin)
	}
	// fmt.Println(exch.orderbooks["BTC-USD"].bid)
	// fmt.Println(exch.orderbooks["BTC-USD"].IsCoherent())
	// os.Exit(1)
	//ask []coinbasepro.SnapshotEntry, bid []coinbasepro.SnapshotEntry
}
func (ob *Orderbook) initBookGDAX(ask []coinbasepro.SnapshotEntry, bid []coinbasepro.SnapshotEntry) {
	log.Println("initializing gdax book")
	ob.update(-1.0, -1.0, -1.0)
	askVarSafe := make([]Level, len(ask)+2, len(ask)+100)
	bidVarSafe := make([]Level, len(bid)+2, len(bid)+100)
	// ob.ask = make([]Level, len(ask)+2) // from big px to small px
	// ob.bid = make([]Level, len(bid)+2) // from small px to big px
	initialUpdateAsks := []interface{}{}
	initialUpdateBids := []interface{}{}
	//curTime := time.Now()
	for idx := 0; idx < len(ask); idx++ {
		element := ask[idx]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)

		if size == 0 {
			continue
		}
		// if ob.record {
		// 	//s := fmt.Sprintf("%f", 123.456)
		// 	ob.updates = append(ob.updates, [3]float64{1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, curTime)
		// }
		ob.update(price, size, 1.0)
		askVarSafe[len(ask)-1-idx+1] = Level{px: price, sz: size}
		initialUpdateAsks = append(initialUpdateAsks, []string{element.Price, element.Size})
	}
	askVarSafe[0] = Level{px: math.MaxFloat64, sz: 1} //cushion the array with extreme values
	askVarSafe[len(askVarSafe)-1] = Level{px: -1, sz: 1}
	for idx := 0; idx < len(bid); idx++ {
		element := bid[idx]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)

		if size == 0 {
			continue
		}
		// if ob.record {
		// 	//s := fmt.Sprintf("%f", 123.456)
		// 	ob.updates = append(ob.updates, [3]float64{-1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, curTime)
		// }
		ob.update(price, size, -1.0)
		bidVarSafe[len(bid)-1-idx+1] = Level{px: price, sz: size}
		initialUpdateBids = append(initialUpdateBids, []string{element.Price, element.Size})
	}
	//SIMULATION CHANGE
	//ob.obUpdates = append(ob.obUpdates, ObUpdate{TimeReceived: curTime, Asks: initialUpdateAsks, Bids: initialUpdateBids})
	bidVarSafe[0] = Level{px: -1, sz: 1} //cushion the array with extreme values
	bidVarSafe[len(bidVarSafe)-1] = Level{px: math.MaxFloat64, sz: 1}
	ob.ask = askVarSafe
	ob.bid = bidVarSafe
	ob.update(-1.0, -1.0, -1.0)
	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}
}

func (ob *Orderbook) updateBookGDAX(change []coinbasepro.SnapshotChange) {
	side := change[0].Side

	if side == "sell" {
		ask_ptr := len(ob.ask) - 1
		lowest_ask_px, _ := strconv.ParseFloat(change[0].Price, 64)
		element := change[0]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)

		for ask_ptr >= 0 && ob.ask[ask_ptr].px < price {
			ask_ptr -= 1
		}
		//cases: level wiped, level adjusted, level inserted
		if ob.ask[ask_ptr].px == price {
			//level wiped
			if size == 0 {
				if ob.CancelFlag {
					ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
						Side:          1,
						Price:         price,
						RemoveSize:    ob.ask[ask_ptr].sz,
						RemainingSize: size,
						Level:         len(ob.ask) - 1 - ask_ptr})
				}

				ob.ask = Remove(ob.ask, ask_ptr)
			} else { //level adjusted
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
			}
		} else { //level inserted
			//fmt.Println("Insert")
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
			}
		}
		// if ob.record {
		// 	ob.updates = append(ob.updates, [3]float64{1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, time.Now())
		// }
		ob.update(price, size, 1.0)
		for ob.bid[len(ob.bid)-2].px >= lowest_ask_px {
			//fmt.Println(ob.bid, lowest_ask_px)
			//fmt.Println(ob.exchange + " CROSSED")
			ob.bid = Remove(ob.bid, len(ob.bid)-2)
		}
	} else {
		highest_bid_px, _ := strconv.ParseFloat(change[0].Price, 64)
		bid_ptr := len(ob.bid) - 1
		element := change[0]
		price, _ := strconv.ParseFloat(element.Price, 64)
		size, _ := strconv.ParseFloat(element.Size, 64)

		for bid_ptr >= 0 && ob.bid[bid_ptr].px > price {
			bid_ptr -= 1
		}
		//cases: level wiped, level adjusted, level inserted
		if ob.bid[bid_ptr].px == price {
			//level wiped
			if size == 0 {
				if ob.CancelFlag {
					ob.RecentCancels.PushBack(RecentCancel{timestamp: time.Now(),
						Side:          -1,
						Price:         price,
						RemoveSize:    ob.bid[bid_ptr].sz,
						RemainingSize: size,
						Level:         len(ob.bid) - 1 - bid_ptr})
				}

				//fmt.Println("bid remove")
				ob.bid = Remove(ob.bid, bid_ptr)
			} else { //level adjusted
				//fmt.Println("bid adjust")
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
			}
		} else { //level inserted
			//fmt.Println("bid insert")
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
			}
		}
		for ob.ask[len(ob.ask)-2].px <= highest_bid_px {
			//fmt.Println(ob.ask, highest_bid_px)
			//fmt.Println(ob.exchange + " CROSSED")
			ob.ask = Remove(ob.ask, len(ob.ask)-2)
		}
		// if ob.record {
		// 	ob.updates = append(ob.updates, [3]float64{-1.0, price, size})
		// 	ob.updateTimes = append(ob.updateTimes, time.Now())
		// }
		ob.update(price, size, -1.0)
	}
	if ob.PrevMidFlag && ob.curMid != ob.Mid() {
		ob.curMid = ob.Mid()
		ob.prevMids.PushBack(PriorMid{timestamp: time.Now(), mid: ob.curMid})
	}

}
func (exch *GDAXExchange) readMessages(done chan struct{}, c *websocket.Conn) {
	defer close(done)
	//init_flag := false
	for {
		message := coinbasepro.Message{}
		if err := c.ReadJSON(&message); err != nil {
			log.Println("gdax message failure")
			println(err.Error())
			break
		}
		//println(message.Type)
		// println(message.Asks)
		// println(message.Bids)
		coin := message.ProductID
		msg_type := message.Type
		if msg_type == "snapshot" {
			exch.orderbooks[coin].initBookGDAX(message.Asks, message.Bids)
		} else if msg_type == "l2update" {
			//fmt.Println("gdax l2 update", coin, message.Changes)
			exch.orderbooks[coin].updateBookGDAX(message.Changes)
			atomic.AddUint64(exch.ctrs[coin], uint64(1))
		} else if msg_type == "last_match" {
			exch.trades[coin].updateGDAXTrades(message)
		} else if msg_type == "match" {
			exch.trades[coin].updateGDAXTrades(message)
			atomic.AddUint64(exch.ctrs[coin], uint64(1))
		}
	}
}

func (exch *GDAXExchange) connectCoin() {
	for {
		log.Println("gdax connect")
		client := coinbasepro.NewClient()
		client.RetryCount = 3
		// optional, configuration can be updated with ClientConfig
		client.UpdateConfig(&coinbasepro.ClientConfig{
			BaseURL:    "https://api.pro.coinbase.com",
			Key:        "OsKfdWmBwubcIFHi",
			Passphrase: "noSpoofing_123@%",
			Secret:     "aQecdQgFz9NqGOAEtnSM53TN0cdNG6Za",
		})
		var wsDialer websocket.Dialer
		wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		log.Println("gdax websocket connect orderbook/trades")
		subscribe := coinbasepro.Message{
			Type: "subscribe",
			Channels: []coinbasepro.MessageChannel{
				coinbasepro.MessageChannel{
					Name:       "heartbeat",
					ProductIds: exch.SymbolList,
				},
				coinbasepro.MessageChannel{
					Name:       "level2",
					ProductIds: exch.SymbolList,
				},
				coinbasepro.MessageChannel{
					Name:       "matches",
					ProductIds: exch.SymbolList,
				},
			},
		}
		done := make(chan struct{})
		if err := wsConn.WriteJSON(subscribe); err != nil {
			println(err.Error())
			os.Exit(1)
		}
		go exch.readMessages(done, wsConn)
	L:
		for {
			select {
			case <-done:
				log.Println("gdax reader failure")
				break L
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (exch *GDAXExchange) ConnectFeed() {
	log.Println("GDAX Connect Feed")
	go exch.connectCoin()
	fmt.Println("GDAX Initialized")
}
func (exch *GDAXExchange) SetRecord(mode bool) {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exch.orderbooks[coin].record = mode
		exch.trades[coin].record = mode
	}
}

func (exch *GDAXExchange) DumpRecord(dir string) {
	//permissions := 0644 // or whatever you need
	// err := os.WriteFile("debug.txt", exch.byteArray, 0666)
	// if err != nil {
	// 	// handle error
	// }
	// os.Exit(1)
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		exchPath := CarvePathDate(dir, strings.Replace(coin, "/", "-", 3), exch.name)
		exch.orderbooks[coin].WriteCSV(exchPath)
		exch.trades[coin].WriteCSV(exchPath)
	}
	// for i := 0; i < exch.orderbooks["BTC-USD"].RecentAdds.Len(); i++ {
	// 	fmt.Println(exch.orderbooks["BTC-USD"].RecentAdds.At(i).timestamp,
	// 		exch.orderbooks["BTC-USD"].RecentAdds.At(i).AddSize,
	// 		exch.orderbooks["BTC-USD"].RecentAdds.At(i).Level,
	// 		exch.orderbooks["BTC-USD"].RecentAdds.At(i).OrigSize,
	// 		exch.orderbooks["BTC-USD"].RecentAdds.At(i).Price,
	// 		exch.orderbooks["BTC-USD"].RecentAdds.At(i).Side)
	// }
	// os.Exit(2)

}
func (exch *GDAXExchange) DumpLengths() {
	for i := range exch.SymbolList {
		coin := exch.SymbolList[i]
		fmt.Println(len(exch.orderbooks[coin].updates))
		//fmt.Println(len(exch.trades[coin].updates))
	}
}
