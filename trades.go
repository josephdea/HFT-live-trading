package main

import (
	"encoding/csv"
	"fmt"

	//"math"
	"os"
	"strings"
	"time"

	"github.com/gammazero/deque"
)

type Trade struct {
	CreateAt time.Time `json:"createAt"`
	//Liquidation bool      `json:"liquidation"`
	Px   float64 `json:"price"`
	Side string  `json:"side"`
	Sz   float64 `json:"size"`
}
type TradeInfo struct {
	timestamp time.Time
	price     float64
	side      float64
	size      float64
}
type Trades struct {
	window       time.Duration
	limit        int
	recentTrades []Trade
	//recentTrades *deque.Deque[Trade] //data for predictions during time of query
	messageID      int
	updates        [][]float64
	updateTimes    []time.Time
	record         bool
	writingctr     int
	TradeFlag      bool
	recentTradesv1 deque.Deque[TradeInfo]

	// recentTradesv2 deque.Deque[Trade1]
}

func NewTrades() *Trades {
	t := Trades{
		messageID:    -1,
		limit:        1000,
		recentTrades: make([]Trade, 0, 1000),
		updates:      make([][]float64, 0),
		updateTimes:  make([]time.Time, 0),
		record:       false,
		writingctr:   1,
		TradeFlag:    false,
	}
	return &t
}
func (t *Trades) RemoveTrade() {
	t.recentTrades = t.recentTrades[1:]
}
func (t *Trades) Update(timestamp time.Time, price float64, side string, size float64) {
	if t.record == false {
		return
	}
	//fmt.Println(timestamp)
	t_obj := Trade{timestamp, price, side, size}

	t.recentTrades = append(t.recentTrades, t_obj)
	if len(t.recentTrades) > t.limit {
		t.RemoveTrade()
	}
	//strings.ToLower(side)
	sideFloat := 1.0
	if strings.Contains(strings.ToLower(side), "b") {
		sideFloat = -1
	}
	if t.TradeFlag {
		t.recentTradesv1.PushBack(TradeInfo{timestamp: timestamp, price: price, side: sideFloat, size: size})
	}
	//fmt.Println("TRADE")
	if t.record {

		t.updates = append(t.updates, []float64{price, size, sideFloat})
		t.updateTimes = append(t.updateTimes, timestamp)
	}

}
func (t *Trades) WriteCSV(path string) {
	if t.record == false {
		return
	}
	csvFile, _ := os.Create(path + "/" + "trades" + fmt.Sprint(t.writingctr) + ".csv")
	t.writingctr += 1
	csvwriter := csv.NewWriter(csvFile)
	curLength := len(t.updateTimes)
	//fmt.Println(len(t.updates), len(t.updateTimes))
	_ = csvwriter.Write([]string{"Time Received", "Price", "Size", "Side"})
	loc, _ := time.LoadLocation("America/Chicago")
	for i := 0; i < curLength; i++ {
		//strRecord := FloatArrtoStringArr(record)
		line := append([]string{t.updateTimes[i].In(loc).String()}, FloatArrtoStringArr(t.updates[i][:])...)
		_ = csvwriter.Write(line)
		t.updates[i] = nil
	}
	t.updates = t.updates[curLength:]
	t.updateTimes = t.updateTimes[curLength:]
	csvwriter.Flush()
	csvFile.Close()

}
