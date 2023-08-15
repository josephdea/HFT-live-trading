package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

type CancelSignalMid struct {
	nexus       *Nexus
	exchange    string
	coin        string
	book        *Orderbook
	c_size      []float64
	r_size      []float64
	levelcap    []float64
	horizon     []float64
	horizon_idx map[float64]int
	max_horizon *float64
	num_signals int
	write_idx   int
}

func NewCancelSignalMid(n *Nexus, exchange string, coin string, w_idx int, mh *float64) *CancelSignalMid {
	s := CancelSignalMid{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.book.CancelFlag = true
	s.coin = coin
	s.exchange = exchange
	s.c_size = make([]float64, 0)
	s.r_size = make([]float64, 0)
	s.levelcap = make([]float64, 0)
	s.horizon = make([]float64, 0)
	s.horizon_idx = make(map[float64]int)
	s.num_signals = 0
	s.write_idx = w_idx
	s.max_horizon = mh
	return &s
}

func (s *CancelSignalMid) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64

	val, ok := param["c_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_size for CancelSignalMid")
		os.Exit(1)
	}
	s.c_size = append(s.c_size, val_f)
	val, ok = param["r_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no r_size for CancelSignalMid")
		os.Exit(1)
	}
	s.r_size = append(s.r_size, val_f)
	val, ok = param["levelcap"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no levelcap for CancelSignalMid")
		os.Exit(1)
	}
	s.levelcap = append(s.levelcap, val_f)
	val, ok = param["horizon"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no horizon for CancelSignalMid")
		os.Exit(1)
	}
	//s.max_horizon = math.Max(s.max_horizon, val_f)
	s.horizon = append(s.horizon, val_f)

}
func (s *CancelSignalMid) binary_search_upper(t time.Time) int {
	left := 0
	right := s.book.RecentCancels.Len() - 1
	res := right + 1
	for left <= right {
		mid := (left + right) / 2
		if t.Before(s.book.RecentCancels.At(mid).timestamp) {
			res = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return res
}
func (s *CancelSignalMid) preprocess() {
	cur_time := time.Now()
	idx := s.binary_search_upper(cur_time.Add(-time.Duration(*s.max_horizon*1000) * time.Millisecond))

	for i := 0; i < idx; i++ {
		s.book.RecentCancels.PopFront()
	}
	for _, val := range s.horizon {
		s.horizon_idx[val] = s.binary_search_upper(cur_time.Add(-time.Duration(val*1000) * time.Millisecond))
	}

}

func (s *CancelSignalMid) forward() {
	curmid := s.book.Mid()
	for i := 0; i < len(s.c_size); i++ {
		cs := s.c_size[i]
		rs := s.r_size[i]
		lc := s.levelcap[i]
		horizon := s.horizon[i]
		j := s.write_idx + i
		volume := 0.0
		numerator := 0.0
		endidx := s.book.RecentCancels.Len() - 1
		startidx := s.horizon_idx[horizon]

		for k := startidx; k < endidx; k++ {
			price := s.book.RecentCancels.At(k).Price
			lvl := s.book.RecentCancels.At(k).Level
			rsize := s.book.RecentCancels.At(k).RemainingSize
			csize := s.book.RecentCancels.At(k).RemoveSize

			if (lvl >= 2 && lvl <= int(lc)) || (lvl == 1 && lc == 1) {
				effective_volume := csize * math.Exp(-csize*cs) * math.Exp(-rsize*rs)
				numerator += effective_volume * price
				volume += effective_volume
			}

		}

		var res float64
		res = ((numerator / volume) - curmid) / curmid
		if math.IsNaN(res) || math.IsInf(res, 0) {
			res = 0
		}
		s.nexus.SignalOut[j] = res
	}
}
