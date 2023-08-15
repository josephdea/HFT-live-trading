package main

import (
	"fmt"
	"strconv"
)

type StupidOrdex struct {
	book   *Orderbook
	signal func() float64
	//tradeExecutor Executor
	//AskOrders
	//BidOrders
	//available notional
	//signals
	Exec ExecutorStruct
	//TradeConnector interface, can call place order/cancel order/get active order/get active position
	SpreadCoefficient float64
	MidLinearBias     float64
}

func NewStupidOrdex(nexus *Nexus, Args map[string]string) *StupidOrdex {
	ordex := StupidOrdex{}
	exchange := Args["Exchange"]
	symbol := Args["Coin"]
	signalName := Args["Signal Name"]
	ordex.book = nexus.GetBook(exchange, symbol)
	ordex.signal = nexus.Signals[signalName]
	ordex.SpreadCoefficient, _ = strconv.ParseFloat(Args["SpreadCoefficient"], 64)
	ordex.MidLinearBias, _ = strconv.ParseFloat(Args["MidLinearBias"], 64)
	//ordex.tradeExecutor = nexus.GetExecutor(exchange)
	return &ordex
}

func (ordex *StupidOrdex) DoTrade() {
	fmt.Println("Stupid Ordex doing trade")
	pred := ordex.signal()
	fmt.Println(pred)

}

//connect StupidOrdex to the Nexud during Initialization
//func Initialize

//func DoTrade
