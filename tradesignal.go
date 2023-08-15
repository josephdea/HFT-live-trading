package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

type TradeSignal struct {
	nexus         *Nexus
	exchange      string
	coin          string
	trade_obj     *Trades
	book          *Orderbook
	c_size        []float64
	c_prune       []float64
	c_px          []float64
	norm          []float64
	starthorizon  []float64
	endhorizon    []float64
	uniquehorizon []float64
	sizeflag      []bool
	horizon_idx   map[float64]int
	max_horizon   float64
	num_signals   int
	write_idx     int
}

func NewTradeSignal(n *Nexus, exchange string, coin string, w_idx int) *TradeSignal {
	s := TradeSignal{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.trade_obj = n.GetTrade(exchange, coin)
	s.trade_obj.TradeFlag = true
	s.coin = coin
	s.exchange = exchange
	s.c_size = make([]float64, 0)
	s.c_prune = make([]float64, 0)
	s.c_px = make([]float64, 0)
	s.norm = make([]float64, 0)
	s.starthorizon = make([]float64, 0)
	s.endhorizon = make([]float64, 0)
	s.horizon_idx = make(map[float64]int)
	s.num_signals = 0
	s.write_idx = w_idx
	s.max_horizon = 0
	return &s
}

//	func removeDuplicate[T string | int](sliceList []T) []T {
//	    allKeys := make(map[T]bool)
//	    list := []T{}
//	    for _, item := range sliceList {
//	        if _, value := allKeys[item]; !value {
//	            allKeys[item] = true
//	            list = append(list, item)
//	        }
//	    }
//	    return list
//	}
func (s *TradeSignal) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64
	val, ok := param["c_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_size for TradeSignal")
		os.Exit(1)
	}
	s.c_size = append(s.c_size, val_f)

	val, ok = param["c_prune"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_prune for TradeSignal")
		os.Exit(1)
	}
	s.c_prune = append(s.c_prune, val_f)

	val, ok = param["c_px"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_px for TradeSignal")
		os.Exit(1)
	}
	s.c_px = append(s.c_px, val_f)

	val, ok = param["norm"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no norm for TradeSignal")
		os.Exit(1)
	}
	s.norm = append(s.norm, val_f)

	val, ok = param["size_flag"]
	val_b, _ := strconv.ParseBool(val[0])
	if !ok {
		fmt.Println("no size_flag for TradeSignal")
		os.Exit(1)
	}
	s.sizeflag = append(s.sizeflag, val_b)

	val, ok = param["starthorizon"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no starthorizon for TradeSignal")
		os.Exit(1)
	}
	s.starthorizon = append(s.starthorizon, val_f)
	s.uniquehorizon = append(s.uniquehorizon, val_f)
	val, ok = param["endhorizon"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no endhorizon for TradeSignal")
		os.Exit(1)
	}
	s.endhorizon = append(s.endhorizon, val_f)
	s.max_horizon = math.Max(s.max_horizon, val_f)
	s.uniquehorizon = append(s.uniquehorizon, val_f)

	s.uniquehorizon = removeDuplicate(s.uniquehorizon)
	sort.Float64s(s.uniquehorizon)

}

func (s *TradeSignal) binary_search_upper(t time.Time) int {
	left := 0
	right := s.trade_obj.recentTradesv1.Len() - 1
	res := right + 1
	for left <= right {
		mid := (left + right) / 2
		if t.Before(s.trade_obj.recentTradesv1.At(mid).timestamp) {
			res = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return res
}
func (s *TradeSignal) preprocess() {
	//TradeSignal_GDAX_BTC-USD_1e-1_0_0_3_1_0.000000_1.000000

	cur_time := time.Now()
	idx := s.binary_search_upper(cur_time.Add(-time.Duration(s.max_horizon*1000) * time.Millisecond))

	for i := 0; i < idx; i++ {
		s.trade_obj.recentTradesv1.PopFront()
	}
	for _, val := range s.uniquehorizon {

		s.horizon_idx[val] = s.binary_search_upper(cur_time.Add(-time.Duration(val*1000) * time.Millisecond))
	}

}

func (s *TradeSignal) forward() {
	curmid := s.book.Mid()
	// if s.coin == "BTC-USD" && s.exchange == "GDAX" {
	// 	fmt.Println("forward call")
	// }
	for i := 0; i < len(s.c_size); i++ {
		// if s.coin == "BTC-USD" && s.exchange == "GDAX" {
		// 	fmt.Println("wtf i", i, s.c_size)
		// }
		cs := s.c_size[i]
		ct := s.c_prune[i]
		cp := s.c_px[i]
		norm := s.norm[i]
		startHorizon := s.starthorizon[i]
		endHorizon := s.endhorizon[i]
		size_flag := s.sizeflag[i]
		j := s.write_idx + i
		endidx := s.horizon_idx[endHorizon]
		startidx := s.horizon_idx[startHorizon]
		askStrength := 0.0
		bidStrength := 0.0
		volume := 0.0

		for k := endidx; k < startidx; k++ {
			side := s.trade_obj.recentTradesv1.At(k).side
			size := s.trade_obj.recentTradesv1.At(k).size
			if size_flag {
				size = math.Pow(size, cs)
			} else {
				size = size * math.Exp(-size*cs)
			}
			price := s.trade_obj.recentTradesv1.At(k).price
			pricediff := math.Abs(price - curmid)
			value := size * math.Exp(-cp*pricediff) //test different weightings
			//std::cout << value << " " << size << " " << pricediff << " " << price << " " << curmid << " " << RecentTrades.size() << std::endl;
			volume += value
			if ct == 1 {
				if side == 1 && price >= curmid {
					askStrength += value
				}
				if side == -1 && price <= curmid {
					bidStrength += value
				}
			} else {
				if side == 1.0 {
					askStrength += value
				} else {
					bidStrength += value
				}
			}
		}
		var res float64
		if norm == 1 {
			res = math.Log(askStrength+1e-8) - math.Log(bidStrength+1e-8)
		} else if norm == 2 {
			res = (askStrength - bidStrength) / (volume + 1e-8)
		} else if norm == 3 {
			res = askStrength - bidStrength
		}
		if math.IsNaN(res) || math.IsInf(res, 0) {
			res = 0
		}

		s.nexus.SignalOut[j] = res

	}

}
