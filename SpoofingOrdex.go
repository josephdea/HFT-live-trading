package main

import (
	"strconv"
)

type SpoofingOrdex struct {
	book              *Orderbook
	signal            func() float64
	Exec              ExecutorStruct
	SpreadCoefficient float64
	MidLinearBias     float64
}

func NewSpoofingOrdex(nexus *Nexus, Args map[string]string) *SpoofingOrdex {
	ordex := SpoofingOrdex{}
	exchange := Args["Exchange"]
	symbol := Args["Coin"]
	signalName := Args["Signal Name"]
	ordex.book = nexus.GetBook(exchange, symbol)
	ordex.signal = nexus.Signals[signalName]
	ordex.SpreadCoefficient, _ = strconv.ParseFloat(Args["SpreadCoefficient"], 64)
	ordex.MidLinearBias, _ = strconv.ParseFloat(Args["MidLinearBias"], 64)
	return &ordex
}

func (ordex *SpoofingOrdex) DoTrade() {
	// fmt.Println("Stupid Ordex doing trade", time.Now())
	// //ordex.Exec.CancelAllOrders()
	// //mid := (ordex.book.ask[len(ordex.book.ask)-2].px + ordex.book.bid[len(ordex.book.bid)-2].px) * 0.5
	// pred := ordex.signal()
	// pred = 26750
	// fmt.Println("quoting mid at ", pred)
	// curSpread := 150 * ordex.SpreadCoefficient
	// ordex.Exec.SendALO("BUY", pred-0.5*curSpread, 0.001)
	// ordex.Exec.SendALO("SELL", pred+0.5*curSpread, 0.001)
	//fmt.Println(pred)
	// PredMid := (1 + pred) * mid
	// curSpread := ordex.book.ask[len(ordex.book.ask)-2].px - ordex.book.bid[len(ordex.book.bid)-2].px

}
