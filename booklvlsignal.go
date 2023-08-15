package main

import (
	"fmt"
	"os"
	"strconv"
)

type BooklvlSignal struct {
	nexus        *Nexus
	exchange     string
	coin         string
	book         *Orderbook
	level_start  []float64
	levels       []float64
	mid_exchange []string
	c            []float64
	e            []float64
	num_signals  int
	write_idx    int
}

func NewBooklvlSignal(n *Nexus, exchange string, coin string, w_idx int) *BooklvlSignal {
	s := BooklvlSignal{}
	s.nexus = n
	s.book = n.GetBook(exchange, coin)

	s.coin = coin
	s.exchange = exchange

	s.level_start = make([]float64, 0)
	s.levels = make([]float64, 0)
	s.mid_exchange = make([]string, 0)
	s.c = make([]float64, 0)
	s.e = make([]float64, 0)

	s.num_signals = 0
	s.write_idx = w_idx

	return &s
}

func (s *BooklvlSignal) initParam(param map[string][]string) {
	s.num_signals += 1
	var val_f float64
	val, ok := param["level_start"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no level_start for TradeSignal")
		os.Exit(1)
	}
	s.level_start = append(s.level_start, val_f)

	val, ok = param["levels"]
	val_f, _ = strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no levels for TradeSignal")
		os.Exit(1)
	}
	s.levels = append(s.levels, val_f)

	val, ok = param["mid_exchange"]
	val_s := val[0]
	if !ok {
		fmt.Println("no mid_exchange for TradeSignal")
		os.Exit(1)
	}
	s.mid_exchange = append(s.mid_exchange, val_s)

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

func (s *BooklvlSignal) preprocess() {
	return
}

func (s *BooklvlSignal) forward() {
	for i := 0; i < len(s.level_start); i++ {
		lvl_start := s.level_start[i]
		lvl_cap := s.levels[i]
		c := s.c[i]
		e := s.e[i]
		j := s.write_idx + i
		midex := s.mid_exchange[i]
		askNum, askDen := s.book.TraverseSidePxSzLevels(int(lvl_start), 1, int(lvl_cap), c, e)
		bidNum, bidDen := s.book.TraverseSidePxSzLevels(int(lvl_start), -1, int(lvl_cap), c, e)
		curmid := s.nexus.GetBook(midex, s.coin).Mid()
		newmid := 0.5*(askNum/askDen) + 0.5*(bidNum/bidDen)
		s.nexus.SignalOut[j] = (newmid - curmid) / curmid
	}

}
