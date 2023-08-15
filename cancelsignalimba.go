package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

type CancelSignalImba struct {
	nexus       *Nexus
	exchange    string
	coin        string
	book        *Orderbook
	c_lvl       []float64
	c_size      []float64
	r_size      []float64
	levelcap    []float64
	horizon     []float64
	norm        []float64
	horizon_idx map[float64]int
	max_horizon *float64
	num_signals int
	write_idx   int
}

func NewCancelSignalImba(n *Nexus, exchange string, coin string, w_idx int, mh *float64) *CancelSignalImba {
	s := CancelSignalImba{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.book.CancelFlag = true
	s.coin = coin
	s.exchange = exchange
	s.c_lvl = make([]float64, 0)
	s.c_size = make([]float64, 0)
	s.r_size = make([]float64, 0)
	s.levelcap = make([]float64, 0)
	s.horizon = make([]float64, 0)
	s.norm = make([]float64, 0)
	s.horizon_idx = make(map[float64]int)
	s.num_signals = 0
	s.write_idx = w_idx
	s.max_horizon = mh
	return &s
}

func (s *CancelSignalImba) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64
	val, ok := param["c_lvl"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_lvl for CancelSignalImba")
		os.Exit(1)
	}
	s.c_lvl = append(s.c_lvl, val_f)
	val, ok = param["c_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c_size for CancelSignalImba")
		os.Exit(1)
	}
	s.c_size = append(s.c_size, val_f)
	val, ok = param["r_size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no r_size for CancelSignalImba")
		os.Exit(1)
	}
	s.r_size = append(s.r_size, val_f)
	val, ok = param["levelcap"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no levelcap for CancelSignalImba")
		os.Exit(1)
	}
	s.levelcap = append(s.levelcap, val_f)
	val, ok = param["horizon"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no horizon for CancelSignalImba")
		os.Exit(1)
	}
	//s.max_horizon = math.Max(s.max_horizon, val_f)
	s.horizon = append(s.horizon, val_f)
	val, ok = param["norm"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no norm for CancelSignalImba")
		os.Exit(1)
	}
	s.norm = append(s.norm, val_f)

}

func (s *CancelSignalImba) binary_search_upper(t time.Time) int {
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

func (s *CancelSignalImba) preprocess() {
	cur_time := time.Now()
	idx := s.binary_search_upper(cur_time.Add(-time.Duration(*s.max_horizon*1000) * time.Millisecond))

	for i := 0; i < idx; i++ {
		s.book.RecentCancels.PopFront()
	}
	for _, val := range s.horizon {
		s.horizon_idx[val] = s.binary_search_upper(cur_time.Add(-time.Duration(val*1000) * time.Millisecond))
	}

}

func (s *CancelSignalImba) forward() {
	curmid := s.book.Mid()
	for i := 0; i < len(s.norm); i++ {
		norm := s.norm[i]
		cl := s.c_lvl[i]
		cs := s.c_size[i]
		rs := s.r_size[i]
		lc := s.levelcap[i]
		horizon := s.horizon[i]
		j := s.write_idx + i
		askStrength := 0.0
		bidStrength := 0.0
		endidx := s.book.RecentCancels.Len() - 1
		startidx := s.horizon_idx[horizon]

		for k := startidx; k < endidx; k++ {
			side := s.book.RecentCancels.At(k).Side
			price := s.book.RecentCancels.At(k).Price
			//price_diff := math.Abs(curmid - price)
			price_diff := math.Abs(curmid - price)
			lvl := s.book.RecentCancels.At(k).Level
			rsize := s.book.RecentCancels.At(k).RemainingSize
			csize := s.book.RecentCancels.At(k).RemoveSize
			if (lvl >= 2 && lvl <= int(lc)) || (lvl == 1 && lc == 1) {
				value := math.Pow(rsize, rs) * math.Pow(csize, cs) * math.Exp(-cl*price_diff)
				if side == 1 {
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
