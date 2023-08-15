package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

type ReversionSignal struct {
	nexus         *Nexus
	exchange      string
	coin          string
	book          *Orderbook
	starthorizon  []float64
	endhorizon    []float64
	uniquehorizon []float64
	sizeflag      []bool
	horizon_idx   map[float64]int
	max_horizon   float64
	num_signals   int
	write_idx     int
}

func NewReversionSignal(n *Nexus, exchange string, coin string, w_idx int) *ReversionSignal {
	s := ReversionSignal{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.book.PrevMidFlag = true
	s.coin = coin
	s.exchange = exchange
	s.starthorizon = make([]float64, 0)
	s.endhorizon = make([]float64, 0)
	s.horizon_idx = make(map[float64]int)
	s.num_signals = 0
	s.write_idx = w_idx
	s.max_horizon = 0
	return &s
}

func (s *ReversionSignal) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64

	val, ok := param["starthorizon"]
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
func (s *ReversionSignal) binary_search_lower(t time.Time) int {
	left := 0
	right := s.book.prevMids.Len() - 1
	res := left
	for left <= right {
		mid := (left + right) / 2
		//if target < mid
		if t.Before(s.book.prevMids.At(mid).timestamp) {

			right = mid - 1

		} else {
			res = mid
			left = mid + 1

		}
	}
	return res
}
func (s *ReversionSignal) preprocess() {

	cur_time := time.Now()

	idx := s.binary_search_lower(cur_time.Add(-time.Duration(s.max_horizon*1000) * time.Millisecond))

	for i := 0; i < idx; i++ {
		s.book.prevMids.PopFront()
	}

	for _, val := range s.uniquehorizon {
		s.horizon_idx[val] = s.binary_search_lower(cur_time.Add(-time.Duration(val*1000) * time.Millisecond))

	}

}

func (s *ReversionSignal) forward() {
	for i := 0; i < len(s.starthorizon); i++ {
		sh := s.starthorizon[i]
		eh := s.endhorizon[i]
		j := s.write_idx + i
		sh_idx := s.horizon_idx[sh]
		eh_idx := s.horizon_idx[eh]
		sh_mid := s.book.prevMids.At(sh_idx).mid
		eh_mid := s.book.prevMids.At(eh_idx).mid
		returns := (sh_mid - eh_mid) / eh_mid
		if math.IsNaN(returns) || math.IsInf(returns, 0) {
			returns = 0
		}
		s.nexus.SignalOut[j] = returns

	}
}
