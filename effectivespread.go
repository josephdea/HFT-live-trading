package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

type EffectiveSpread struct {
	nexus       *Nexus
	exchange    string
	coin        string
	book        *Orderbook
	level_start []float64
	size        []float64
	c           []float64
	e           []float64
	num_signals int
	write_idx   int
}

func NewEffectiveSpread(n *Nexus, exchange string, coin string, w_idx int) *EffectiveSpread {
	s := EffectiveSpread{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)
	s.coin = coin
	s.exchange = exchange
	s.level_start = make([]float64, 0)
	s.size = make([]float64, 0)
	s.c = make([]float64, 0)
	s.e = make([]float64, 0)
	s.num_signals = 0
	s.write_idx = w_idx
	return &s
}

func (s *EffectiveSpread) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64
	val, ok := param["level_start"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no level_start for TradeSignal")
		os.Exit(1)
	}
	s.level_start = append(s.level_start, val_f)

	val, ok = param["size"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no size for TradeSignal")
		os.Exit(1)
	}
	s.size = append(s.size, val_f)

	val, ok = param["c"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no c for TradeSignal")
		os.Exit(1)
	}
	s.c = append(s.c, val_f)

	val, ok = param["e"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no e for TradeSignal")
		os.Exit(1)
	}
	s.e = append(s.e, val_f)

}

func (s *EffectiveSpread) preprocess() {
	return
}

func (s *EffectiveSpread) forward() {
	for i := 0; i < len(s.level_start); i++ {
		lvl_start := s.level_start[i]
		size := s.size[i]
		c := s.c[i]
		e := s.e[i]
		j := s.write_idx + i
		askNum, askDen := s.book.TraverseSidePxSzSize(int(lvl_start), 1, size, c, e)

		bidNum, bidDen := s.book.TraverseSidePxSzSize(int(lvl_start), -1, size, c, e)

		spread := (askNum / askDen) - (bidNum / bidDen)
		if math.IsNaN(spread) || math.IsInf(spread, 0) {
			spread = 0
		}
		s.nexus.SignalOut[j] = spread
	}

}
