package main

import (
	"fmt"
	"os"
	"strconv"
)

type BookszSignal struct {
	nexus        *Nexus
	exchange     string
	coin         string
	book         *Orderbook
	level_start  []float64
	size         []float64
	mid_exchange []string
	c            []float64
	e            []float64
	num_signals  int
	write_idx    int
}

func NewBookszSignal(n *Nexus, exchange string, coin string, w_idx int) *BookszSignal {
	s := BookszSignal{}
	s.nexus = n
	fmt.Println("books sz signal", exchange, coin)

	s.book = n.GetBook(exchange, coin)
	fmt.Println("book signal book address :", exchange, coin, &s.book)
	s.coin = coin
	s.exchange = exchange

	s.level_start = make([]float64, 0)
	s.size = make([]float64, 0)
	s.mid_exchange = make([]string, 0)
	s.c = make([]float64, 0)
	s.e = make([]float64, 0)

	s.num_signals = 0
	s.write_idx = w_idx

	return &s
}

func (s *BookszSignal) initParam(param map[string][]string) {
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

	val, ok = param["mid_exchange"]
	val_s := val[0]
	s.nexus.AddExchangeCoin(val_s, s.coin)
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

func (s *BookszSignal) preprocess() {
	return
}

func (s *BookszSignal) forward() {
	fmt.Println(s.exchange, s.coin)
	for i := 0; i < len(s.level_start); i++ {
		lvl_start := s.level_start[i]
		size := s.size[i]
		c := s.c[i]
		e := s.e[i]
		j := s.write_idx + i
		midex := s.mid_exchange[i]
		askNum, askDen := s.book.TraverseSidePxSzSize(int(lvl_start), 1, size, c, e)
		bidNum, bidDen := s.book.TraverseSidePxSzSize(int(lvl_start), -1, size, c, e)
		curmid := s.nexus.GetBook(midex, s.coin).Mid()
		newmid := 0.5*(askNum/askDen) + 0.5*(bidNum/bidDen)
		s.nexus.SignalOut[j] = (newmid - curmid) / curmid
	}

}
