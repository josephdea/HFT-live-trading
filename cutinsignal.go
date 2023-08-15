package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

type CutinSignal struct {
	nexus       *Nexus
	exchange    string
	coin        string
	book        *Orderbook
	c_size      []float64
	c_cutin     []float64
	norm        []float64
	horizon     []float64
	horizon_idx map[float64]int
	max_horizon float64
	num_signals int
	write_idx   int
}

func NewCutinSignal(n *Nexus, exchange string, coin string, w_idx int) *CutinSignal {
	s := CutinSignal{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.book.CutinFlag = true
	s.coin = coin
	s.exchange = exchange
	s.c_size = make([]float64, 0)
	s.c_cutin = make([]float64, 0)
	s.horizon = make([]float64, 0)
	s.horizon_idx = make(map[float64]int)
	s.num_signals = 0
	s.write_idx = w_idx
	s.max_horizon = 0
	return &s
}

func (s *CutinSignal) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64

	val, ok := param["c_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_size for CancelSignalMid")
		os.Exit(1)
	}
	s.c_size = append(s.c_size, val_f)

	val, ok = param["c_cutin"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_cutin for CancelSignalMid")
		os.Exit(1)
	}
	s.c_cutin = append(s.c_cutin, val_f)

	val, ok = param["norm"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no norm for CancelSignalMid")
		os.Exit(1)
	}
	s.norm = append(s.norm, val_f)

	val, ok = param["horizon"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no horizon for CancelSignalMid")
		os.Exit(1)
	}
	s.max_horizon = math.Max(s.max_horizon, val_f)
	s.horizon = append(s.horizon, val_f)

}

func (s *CutinSignal) binary_search_upper(t time.Time) int {
	left := 0
	right := s.book.RecentCutIns.Len() - 1
	res := right + 1
	for left <= right {
		mid := (left + right) / 2
		if t.Before(s.book.RecentCutIns.At(mid).timestamp) {
			res = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return res
}

func (s *CutinSignal) preprocess() {

	cur_time := time.Now()
	idx := s.binary_search_upper(cur_time.Add(-time.Duration(s.max_horizon*1000) * time.Millisecond))

	for i := 0; i < idx; i++ {
		s.book.RecentCutIns.PopFront()
	}
	for _, val := range s.horizon {
		s.horizon_idx[val] = s.binary_search_upper(cur_time.Add(-time.Duration(val*1000) * time.Millisecond))
	}

}

func (s *CutinSignal) forward() {
	for i := 0; i < len(s.c_size); i++ {
		cs := s.c_size[i]
		cc := s.c_cutin[i]
		norm := s.norm[i]
		horizon := s.horizon[i]
		j := s.write_idx + i

		endidx := s.book.RecentCutIns.Len() - 1
		startidx := s.horizon_idx[horizon]
		askStrength := 0.0
		bidStrength := 0.0
		for k := startidx; k < endidx; k++ {
			side := s.book.RecentCutIns.At(k).side
			effective_sz := math.Pow(s.book.RecentCutIns.At(k).size, cs)
			cutin_amount := s.book.RecentCutIns.At(k).cutin
			if side == 1 {
				askStrength += effective_sz * math.Exp(-cc*cutin_amount)
			} else {
				bidStrength += effective_sz * math.Exp(-cc*cutin_amount)
			}

		}
		var res float64
		if norm == 0 {
			res = math.Log(askStrength+1e-8) - math.Log(bidStrength+1e-8)
		} else if norm == 1 {
			res = (askStrength - bidStrength) / (askStrength + bidStrength + 1e-8)
		} else {
			res = askStrength - bidStrength
		}

		if math.IsNaN(res) || math.IsInf(res, 0) {
			res = 0
		}
		s.nexus.SignalOut[j] = res
	}
}
