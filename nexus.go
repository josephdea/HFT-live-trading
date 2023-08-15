package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Nexus struct {
	Signals            map[string]func() float64
	Ordexes            []Ordex
	DYDX               DYDXExchange
	KRAKEN             KRAKENExchange
	GDAX               GDAXExchange
	BNB                BNBExchange
	AllBooks           map[string]*Orderbook
	AllTrades          map[string]*Trades
	AllSignals         []SignalStruct
	BaseSignals        map[string]SignalStruct
	UniqueSignals      map[string]int
	OrderedSignalNames []string
	Heartbeats         []func(*uint64, int, func())
	ctrsArr            []*uint64
	thresholdsArr      []int
	DoTradeArr         []func()
	SignalOut          []float64
	Predictions        [][]float64
	PredictionTime     []time.Time
}

func HeartBeat(ctr *uint64, threshold int, execute func()) {
	t := uint64(threshold)

	for {
		//fmt.Println("heartbeat", *ctr, t)
		if *ctr >= t {
			*ctr = 0
			execute()
		}
	}
}
func sum(ctrs []*uint64) uint64 {
	t := uint64(0)
	for i := 0; i < len(ctrs); i++ {
		t += *ctrs[i]
	}
	return t
}
func zero_ctrs(ctrs []*uint64) {
	for i := 0; i < len(ctrs); i++ {
		*ctrs[i] = 0
	}
}
func (n *Nexus) CompleteHeartBeat(ctrs []*uint64, threshold int, record bool) {
	t := uint64(threshold)
	for {
		if sum(ctrs) >= t {
			zero_ctrs(ctrs)
			for i := 0; i < len(n.AllSignals); i++ {
				n.AllSignals[i].signal.preprocess()
				n.AllSignals[i].signal.forward()
			}
			if record {
				tmp := make([]float64, 0)
				for _, val := range n.SignalOut {
					tmp = append(tmp, val)
				}
				n.Predictions = append(n.Predictions, tmp)
				n.PredictionTime = append(n.PredictionTime, time.Now())
			}
		}
	}
}
func (n *Nexus) GetBook(exchange string, coin string) *Orderbook {
	switch exchange {
	case "DYDX":
		return n.DYDX.orderbooks[coin]
	case "GDAX":
		return n.GDAX.orderbooks[coin]
	case "KRAKEN":
		return n.KRAKEN.orderbooks[coin]
	case "BNB":
		return n.BNB.orderbooks[coin]
	default:
		return n.DYDX.orderbooks[coin]
	}
	//return n.DYDX.orderbooks[coin]
}
func (n *Nexus) GetTrade(exchange string, coin string) *Trades {
	switch exchange {
	case "DYDX":
		return n.DYDX.trades[coin]
	case "GDAX":
		return n.GDAX.trades[coin]
	case "KRAKEN":
		return n.KRAKEN.trades[coin]
	case "BNB":
		return n.BNB.trades[coin]
	default:
		return n.DYDX.trades[coin]
	}
	//return n.DYDX.orderbooks[coin]
}
func (n *Nexus) AddExchangeCoin(exchange string, coin string) {
	unique_id := exchange + coin
	_, ok := n.AllBooks[unique_id]
	if ok {
		return
	}
	////n.ctrsArr = append(n.ctrsArr, n.DYDX.ctrs[args["Coin"]])
	switch exchange {
	case "DYDX":
		fmt.Println("Adding", exchange, coin)
		n.DYDX.InitCoin(coin)
		n.ctrsArr = append(n.ctrsArr, n.DYDX.ctrs[coin])

		n.AllBooks[unique_id] = n.DYDX.orderbooks[coin]

	case "GDAX":
		fmt.Println("Adding", exchange, coin)
		n.GDAX.InitCoin(coin)
		n.ctrsArr = append(n.ctrsArr, n.GDAX.ctrs[coin])

		n.AllBooks[unique_id] = n.GDAX.orderbooks[coin]
	case "BNB":
		fmt.Println("Adding", exchange, coin)
		n.BNB.InitCoin(coin)
		n.ctrsArr = append(n.ctrsArr, n.BNB.ctrs[coin])
		n.AllBooks[unique_id] = n.BNB.orderbooks[coin]
	}

}
func (n *Nexus) InitLiveTradingFeeds(fileName string) {
	//parse incoming signals config file
	//initialize signals while initializing exchanges
	n.DYDX = *NewDYDXExchange()
	n.GDAX = *NewGDAXExchange()
	n.BNB = *NewBNBExchange()
	n.AllBooks = make(map[string]*Orderbook)
	n.AllSignals = make([]SignalStruct, 0)
	n.BaseSignals = make(map[string]SignalStruct)
	n.UniqueSignals = make(map[string]int)
	n.ctrsArr = make([]*uint64, 0)
	n.OrderedSignalNames = make([]string, 0)
	n.PredictionTime = make([]time.Time, 0)
	n.Predictions = make([][]float64, 0)
	signals := ParseConfig(fileName)

	w_idx := 0
	//var ch *float64
	ch := new(float64)
	*ch = 0.0
	for _, val := range signals {
		//fmt.Println(val.paramName)
		exchange, eflag := val.params["exchange"]
		coin, cflag := val.params["coin"]
		if eflag && cflag {
			n.AddExchangeCoin(exchange[0], coin[0])
		}

		switch val.paramName {
		case "AddSignalImba":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			//a := n.GetBook("BNB", "BTC-USD")
			//b := n.GetBook("BNB", "BTC-USD")
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewAddSignalImba(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}

			//fmt.Println(&(n.BNB.orderbooks["BTC-USD"]))
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1
		case "CancelSignalImba":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewCancelSignalImba(n, exchange[0], coin[0], w_idx, ch)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			tmpstr, _ := val.params["horizon"]
			val_f, _ := strconv.ParseFloat(tmpstr[0], 64)
			// cancel_horizon = math.Max(cancel_horizon, val_f)
			*ch = math.Max(*ch, val_f)
			w_idx += 1

		case "CancelSignalMid":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewCancelSignalMid(n, exchange[0], coin[0], w_idx, ch)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			tmpstr, _ := val.params["horizon"]
			val_f, _ := strconv.ParseFloat(tmpstr[0], 64)
			// cancel_horizon = math.Max(cancel_horizon, val_f)
			*ch = math.Max(*ch, val_f)
			w_idx += 1

		case "CutinSignal":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewCutinSignal(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "BookSignalLevelsMid":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewBooklvlSignal(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "BookSignalSizeMid":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewBookszSignal(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "EffectiveSpread":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewEffectiveSpread(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "ReversionSignal":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewReversionSignal(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "TradeSignal":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewTradeSignal(n, exchange[0], coin[0], w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1

		case "GeometricSignal":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewGeometricSignal(n, w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			w_idx += len(val.params["inc_signal"])
			//n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)

		case "LinearModel":
			_, flag := n.BaseSignals[GetBaseIdentifier(val)]
			if !flag {
				n.BaseSignals[GetBaseIdentifier(val)] = SignalStruct{NewLinearSignal(n, w_idx)}
				n.AllSignals = append(n.AllSignals, n.BaseSignals[GetBaseIdentifier(val)])
			}
			n.OrderedSignalNames = append(n.OrderedSignalNames, GetSignalIdentifier(val))
			n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
			n.BaseSignals[GetBaseIdentifier(val)].signal.initParam(val.params)
			w_idx += 1
		default:
			fmt.Println("Signal not Recognized")
			os.Exit(1)

		}

	}
	*ch += 0.5
	n.SignalOut = make([]float64, w_idx)

}

func (n *Nexus) InitializeFeeds(fileName string) {
	n.Signals = make(map[string]func() float64)
	n.DYDX = *NewDYDXExchange()
	n.KRAKEN = *NewKRAKENExchange()
	n.GDAX = *NewGDAXExchange()
	n.BNB = *NewBNBExchange()
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("coins.txt file does not exist")
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	args := make(map[string]string)
	dupChecker := make(map[string]bool)

	for {
		line := scanner.Text()
		if len(line) != 0 {
			idx := strings.Index(line, ":")
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			args[key] = val
		}
		next := scanner.Scan()
		if !next || len(line) == 0 {
			_, ok := args["exchange"]
			if ok {
				switch args["exchange"] {
				case "DYDX":
					_, dupFlag := dupChecker["DYDX"+args["coin"]]
					if !dupFlag {
						// fmt.Println("HERE", args["coin"])
						// os.Exit(1)

						n.DYDX.InitCoin(args["coin"])
					}
					dupChecker["DYDX"+args["coin"]] = true
				case "BNB":
					_, dupFlag := dupChecker["BNB"+args["coin"]]
					if !dupFlag {
						// fmt.Println("HERE", args["coin"])
						// os.Exit(1)

						n.BNB.InitCoin(args["coin"])
					}
					dupChecker["BNB"+args["coin"]] = true
				case "KRAKEN":
					_, dupFlag := dupChecker["KRAKEN"+args["coin"]]
					if !dupFlag {
						// fmt.Println("HERE", args["coin"])
						// os.Exit(1)

						n.KRAKEN.InitCoin(args["coin"])
					}
					dupChecker["KRAKEN"+args["coin"]] = true
				case "GDAX":
					_, dupFlag := dupChecker["GDAX"+args["coin"]]
					if !dupFlag {
						// fmt.Println("HERE", args["coin"])
						// os.Exit(1)

						n.GDAX.InitCoin(args["coin"])
					}
					dupChecker["GDAX"+args["coin"]] = true
				}
			}

			args = make(map[string]string)
		}
		if !next {
			break
		}
	}

}

func (n *Nexus) BuildComputeGraph(fileName string) []string {
	//registry := map[string]string{"test":"sdfsd"}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("coins.txt file does not exist")
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	args := make(map[string]string)
	SignalArr := make([]string, 0)
	SignalArgs := make(map[string]map[string]string)

	for {
		line := scanner.Text()
		if len(line) != 0 {
			idx := strings.Index(line, ":")
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			args[key] = val
		}
		next := scanner.Scan()
		if !next || len(line) == 0 {
			SignalName, ok := args["Signal Name"]
			if ok {
				delete(args, "Signal Name")
				SignalArr = append(SignalArr, SignalName)
				SignalArgs[SignalName] = args
				args = make(map[string]string)
			}
		}
		if !next {
			break
		}
	}

	edges := CreateEdges(SignalArr, SignalArgs)
	order := TopoSort(edges)
	return order
}
func (n *Nexus) WritePredictions(fileOut string, length int) {
	csvFile, _ := os.Create(fileOut)
	csvwriter := csv.NewWriter(csvFile)
	cols := []string{"time"}
	cols = append(cols, n.OrderedSignalNames...)
	_ = csvwriter.Write(cols)
	for i := 0; i < length; i++ {
		//fmt.Println(n.Predictions[i])
		line := append([]string{n.PredictionTime[i].String()}, FloatArrtoStringArr(n.Predictions[i])...)
		//line := append([]string{n.PredictionTime[i].String(),FloatArrtoStringArr(n.Predictions[i])...)})
		_ = csvwriter.Write(line)
	}
	csvwriter.Flush()
	csvFile.Close()
}

func (n *Nexus) Execute(record_data bool, record_predictions bool, minutes int, queryRate int, signalfile string) { //whether to record prediction values for debugging, and how long to run
	n.InitLiveTradingFeeds(signalfile)

	n.GDAX.SetRecord(record_data)
	n.BNB.SetRecord(record_data)
	n.DYDX.SetRecord(record_data)
	n.DYDX.ConnectFeed()
	n.BNB.ConnectFeed()
	n.GDAX.ConnectFeed()
	fmt.Println("Warming Up 10 seconds")
	time.Sleep(10 * time.Second)
	n.StartOrdexes(queryRate, record_predictions)
	queryTicker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-queryTicker.C:
				n.DYDX.DumpRecord("./data")
				n.GDAX.DumpRecord("./data")
				n.BNB.DumpRecord("./data")
				// nexus.BNB.DumpRecord("./data")
			}
		}
	}()
	time.Sleep(time.Duration(minutes*60) * time.Second)
	//time.Sleep(10 * time.Second)
	predLength := len(n.Predictions)

	n.WritePredictions("predictions.csv", predLength)
	time.Sleep(time.Duration(60) * time.Second)
	os.Exit(1)
	//load signals
	//initialize feeds
	//warm up
	//begin counting
	//make predictions
	//if record, record timestamp and prediction value
	//at end of specified time, dump prediction values
}

func (n *Nexus) InitializeOrdexes(fileName string) {
	fmt.Println("initializing ordexes")
	n.Ordexes = make([]Ordex, 0, 5)
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("coins.txt file does not exist")
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	args := make(map[string]string)
	for {
		line := scanner.Text()
		if len(line) != 0 {
			idx := strings.Index(line, ":")
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			args[key] = val
		}
		next := scanner.Scan()
		if !next || len(line) == 0 {
			var tradeConnection Executor
			ES := NewExecutorStruct()
			exchange := args["Exchange"]
			switch exchange {
			case "BNB":
			default:
				fmt.Println("Ordex Exchange Not Recognized")
				os.Exit(1)
			}
			ES.Exec = tradeConnection
			ordexType := args["Ordex Type"]
			switch ordexType {

			default:
				fmt.Println("Ordex Type Not Recognized")

				os.Exit(1)
			}
			freq, _ := strconv.Atoi(args["Frequency"])
			n.thresholdsArr = append(n.thresholdsArr, freq)
			n.DoTradeArr = append(n.DoTradeArr, n.Ordexes[len(n.Ordexes)-1].DoTrade)
			args = make(map[string]string)
		}
		if !next {
			break
		}
	}
}

func (n *Nexus) StartOrdexes(threshold int, record bool) {
	fmt.Println("Starting Execution")
	go n.CompleteHeartBeat(n.ctrsArr, threshold, record)

}
