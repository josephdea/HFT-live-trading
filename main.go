package main

import (
	"fmt"
	"os"
	"time"
	//"github.com/shopspring/decimal"
)

var ORDTYPE OrderTypeStruct = OrderTypeStruct{ALO: 0, IOC: 1, FOK: 2}
var SIDE SideStruct = SideStruct{BUY: 0, SELL: 1}

type PrivateRequest struct {
	DYDX_SIGNATURE string `json:"DYDX-SIGNATURE"`
	DYDX_API_KEY   string `json:"DYDX-API-KEY"`
	DYDX_TIMESTAMP string `json:"DYDX-TIMESTAMP"`
	DYDX_PASSPHRAE string `json:"DYDX-PASSPHRASE"`
}
type OnboardRequest struct {
	StarkKey   string `json:"starkKey"`
	StarkY     string `json:"starkKeyYCoordinate"`
	EthAddress string `json:"ethereumAddress"`
}

// start := time.Now()

//     r := new(big.Int)
//     fmt.Println(r.Binomial(1000, 10))

// elapsed := time.Since(start)
// log.Printf("Binomial took %s", elapsed)
func main() {
	exec := NewDYDXexecutorv2("BTC-USD")
	fmt.Println(exec.GetOrders())
	exec.CancelAllOrders()
	fmt.Println(exec.GetActiveBuyOrders())

	os.Exit(1)
	// exec.getPosition()
	// os.Exit(1)
	// ParseConfig("./05-30-2023signals10.config")
	// os.Exit(1)
	// exec := NewBNBexecutor("BTCUSDT")
	// exec.init()
	// //exec.SendALO("BUY", 25500.30, 0.001)
	// fmt.Println(exec.GetActiveOrders())
	// exec.CancelAllOrders()
	// fmt.Println(exec.GetActiveOrders())
	// os.Exit(1)
	//recordtickdata := true
	numSeconds := 15 * 60 * 60 //how long to collect feeds in seconds
	//(record_data bool, record_predictions bool, minutes int, queryRate int, signalfile string)
	record_data := true
	record_predictions := true
	minutes := 20
	queryRate := 200
	signalfile := "./debug.config"
	nexus := Nexus{}

	nexus.Execute(record_data, record_predictions, minutes, queryRate, signalfile)
	// nexus.InitLiveTradingFeeds("./debug.config")
	// //nexus.GDAX.SetRecord(recordtickdata)
	// // nexus.DYDX.ConnectFeed()
	// // nexus.BNB.ConnectFeed()
	// // nexus.GDAX.ConnectFeed()

	// //nexus.BNB.ConnectFeed()
	// fmt.Println("Warming Up 1 seconds")
	// time.Sleep(1 * time.Second)
	// nexus.StartOrdexes(200)
	// queryTicker := time.NewTicker(10 * time.Second) //how frequent to dump tick data to csv's
	// go func() {
	// 	for {
	// 		select {
	// 		case <-queryTicker.C:
	// 			// nexus.DYDX.DumpRecord("./data")
	// 			nexus.GDAX.DumpRecord("./data")
	// 			// // nexus.KRAKEN.DumpRecord("./data")
	// 			// nexus.BNB.DumpRecord("./data")
	// 		}
	// 	}
	// }()
	time.Sleep(time.Duration(numSeconds) * time.Second)
	os.Exit(1)
	//nexus.InitializeFeeds("./debugsignals.config") //determine what signals to subscribe to searches for combinations of exchange/coin keywords
	//nexus.BuildComputeGraph("./signals.config")    //build signal computational graph
	//fmt.Println(nexus.Signals)
	//nexus.DYDX.ConnectFeed() //subscribe to websocket channels

	// nexus.KRAKEN.SetRecord(recordtickdata)
	// nexus.DYDX.SetRecord(recordtickdata)

	// nexus.BNB.SetRecord(recordtickdata)

	// nexus.KRAKEN.ConnectFeed()
	// nexus.DYDX.ConnectFeed()
	// nexus.GDAX.ConnectFeed()
	// nexus.BNB.ConnectFeed()
	// fmt.Println("Waiting")
	// //queryTicker := time.NewTicker(10 * time.Second) //how frequent to dump tick data to csv's
	// go func() {
	// 	for {
	// 		select {
	// 		case <-queryTicker.C:
	// 			// nexus.DYDX.DumpRecord("./data")
	// 			// nexus.GDAX.DumpRecord("./data")
	// 			// // nexus.KRAKEN.DumpRecord("./data")
	// 			// nexus.BNB.DumpRecord("./data")
	// 		}
	// 	}
	// }()
	// time.Sleep(time.Duration(numSeconds) * time.Second)
	// nexus.GDAX.DumpRecord("./data")
	// nexus.KRAKEN.DumpRecord("./data")
	// nexus.DYDX.DumpRecord("./data")
	// os.Exit(1)

	// time.Sleep(time.Duration(numSeconds) * time.Second)
	// //nexus.GDAX.DumpRecord("./data")
	// //nexus.KRAKEN.DumpRecord("./data")
	// //nexus.DYDX.DumpRecord("./data")
	// os.Exit(1)

	// nexus.InitializeOrdexes("./ordex.config")
	// time.Sleep(5 * time.Second)
	// nexus.StartOrdexes()
	// time.Sleep(time.Duration(numSeconds) * time.Second)

}
