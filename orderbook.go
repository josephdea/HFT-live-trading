package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gammazero/deque"
)

type Level struct {
	px float64
	sz float64
}
type CutinOrder struct {
	timestamp time.Time
	spread    float64
	cutin     float64
	size      float64
	side      int
}
type RecentAdd struct {
	timestamp time.Time
	Price     float64
	AddSize   float64
	Level     int
	Side      int
	OrigSize  float64
}
type RecentCancel struct {
	timestamp     time.Time
	Price         float64
	RemoveSize    float64
	Level         int
	Side          int
	RemainingSize float64
}
type PriorMid struct {
	timestamp time.Time
	mid       float64
}
type Orderbook struct {
	ask         []Level
	bid         []Level
	updates     [][]float64
	updateTimes []time.Time
	pu          int
	//queryDepth int
	record     bool
	messageID  int                //used to make sure that messages are received in order
	Askoffsets map[float64]uint64 //for dydx only
	Bidoffsets map[float64]uint64
	offsets    map[float64]uint64
	writingctr int

	CutinFlag   bool
	AddFlag     bool
	CancelFlag  bool
	PrevMidFlag bool

	curMid        float64
	RecentCutIns  deque.Deque[CutinOrder]
	RecentAdds    deque.Deque[RecentAdd]
	RecentCancels deque.Deque[RecentCancel]
	prevMids      deque.Deque[PriorMid] //[]PriorMid
	//prevMids2     deque.Deque[PriorMid] //[]PriorMid

}

func NewOrderbook() *Orderbook {
	ob := Orderbook{
		messageID:   -1,
		ask:         make([]Level, 0, 800),
		bid:         make([]Level, 0, 800),
		updates:     make([][]float64, 0),
		updateTimes: make([]time.Time, 0),
		writingctr:  1,
		CutinFlag:   false,
		AddFlag:     false,
		CancelFlag:  false,
		PrevMidFlag: false,
		record:      false}
	return &ob
}

func (ob Orderbook) IsCoherent() bool {
	if len(ob.bid) > 2 && len(ob.ask) > 2 {
		best_ask := ob.ask[len(ob.ask)-2].px
		best_bid := ob.bid[len(ob.bid)-2].px
		if best_bid >= best_ask {
			fmt.Println("Book Crossed")
			return false
		}
	}
	for i := 0; i < len(ob.bid)-1; i++ {
		if ob.bid[i].px >= ob.bid[i+1].px {
			fmt.Println("Bid", ob.bid[i-1].px, ob.bid[i].px, ob.bid[i+1].px)
			fmt.Println(" Relative Ordering Issue")
			return false
		}
		if ob.bid[i].sz == 0 {
			fmt.Println(" Removal Issue")
			return false
		}
	}
	for i := 0; i < len(ob.ask)-1; i++ {
		if ob.ask[i].px <= ob.ask[i+1].px {
			fmt.Println("Ask", ob.ask[i-1].px, ob.ask[i].px, ob.ask[i+1].px)
			fmt.Println(" Relative Ordering Issue")
			return false
		}
		if ob.ask[i].sz == 0 {
			fmt.Println(" Removal Issue")
			return false
		}
	}
	return true
}

func (ob *Orderbook) WriteCSV(path string) {
	if ob.record == false {
		return
	}
	csvFile, _ := os.Create(path + "/" + "orderbook" + fmt.Sprint(ob.writingctr) + ".csv")
	ob.writingctr += 1
	csvwriter := csv.NewWriter(csvFile)
	_ = csvwriter.Write([]string{"Time Received", "Side", "Price", "Size"})
	curLength := len(ob.updateTimes)
	//fmt.Println(ob.updates)
	loc, _ := time.LoadLocation("America/Chicago")
	//fmt.Println("debug", curLength, len(ob.updateTimes), len(ob.updates), ob.writingctr)
	for i := 0; i < curLength; i++ {
		//strRecord := FloatArrtoStringArr(record)
		line := append([]string{ob.updateTimes[i].In(loc).String()}, FloatArrtoStringArr(ob.updates[i][:])...)
		//ob.updateTimes[i] = nil
		ob.updates[i] = nil
		_ = csvwriter.Write(line)
	}
	csvwriter.Flush()
	csvFile.Close()
	ob.updates = ob.updates[curLength:]
	ob.updateTimes = ob.updateTimes[curLength:]

}
func (ob *Orderbook) Print() {

	mid := (ob.ask[len(ob.ask)-2].px + ob.bid[len(ob.bid)-2].px) * 0.5
	fmt.Println(mid)
}
func (ob *Orderbook) Mid() float64 {

	mid := (ob.ask[len(ob.ask)-2].px + ob.bid[len(ob.bid)-2].px) * 0.5
	return mid
}
func (ob *Orderbook) Spread() float64 {

	spr := (ob.ask[len(ob.ask)-2].px - ob.bid[len(ob.bid)-2].px)
	return spr
}
func (ob *Orderbook) BestAsk() float64 {

	return ob.ask[len(ob.ask)-2].px
}

func (ob *Orderbook) BestBid() float64 {

	return ob.bid[len(ob.bid)-2].px
}

func (ob *Orderbook) update(price float64, size float64, side float64) {
	if ob.record {
		a := []float64{side, price, size}
		ob.updates = append(ob.updates, a)
		ob.updateTimes = append(ob.updateTimes, time.Now())
	}
	//fmt.Println(ob.RecentAdds.Len(), ob.RecentCancels.Len(), ob.RecentCutIns.Len())
}

//Unimportant functions below

func exp_function(val float64, c float64, e float64) float64 {
	return math.Exp(-c * math.Pow(val, e))
}

func (ob *Orderbook) TraverseSidePxSzSize(lvl_start int, side int, size float64, c float64, e float64) (float64, float64) {
	cur_lvl := lvl_start
	var weightedSum, denominator, best_offer, sz, px, weights float64
	var bookHalf []Level
	weightedSum = 0
	denominator = 0

	if side == 1 {
		bookHalf = ob.ask
	} else {
		bookHalf = ob.bid
	}
	best_offer = bookHalf[len(bookHalf)-2-cur_lvl].px
	for size > 0 {
		if len(bookHalf)-2-cur_lvl <= 0 || cur_lvl == 20 {
			break
		}
		sz = math.Min(size, bookHalf[len(bookHalf)-2-cur_lvl].sz)
		px = bookHalf[len(bookHalf)-2-cur_lvl].px
		weights = exp_function(math.Abs(px-best_offer), c, e)

		size -= sz
		weightedSum += px * sz * weights
		denominator += weights * sz
		cur_lvl++

	}
	//exit(1);
	return weightedSum, denominator
}

func (ob *Orderbook) TraverseSidePxSzLevels(lvl_start int, side int, lvl_cap int, c float64, e float64) (float64, float64) {
	cur_lvl := lvl_start
	var weightedSum, denominator, best_offer, sz, px, weights float64
	var bookHalf []Level
	weightedSum = 0
	denominator = 0

	if side == 1 {
		bookHalf = ob.ask
	} else {
		bookHalf = ob.bid
	}
	best_offer = bookHalf[len(bookHalf)-2-cur_lvl].px
	for cur_lvl < lvl_cap {
		if len(bookHalf)-2-cur_lvl <= 0 || cur_lvl == 20 {
			break
		}
		sz = bookHalf[len(bookHalf)-2-cur_lvl].sz
		px = bookHalf[len(bookHalf)-2-cur_lvl].px
		weights = exp_function(math.Abs(px-best_offer), c, e)
		weightedSum += px * sz * weights
		denominator += sz * weights
		cur_lvl++
	}
	return weightedSum, denominator
}
