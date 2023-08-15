package main

type MidSignal struct {
	book *Orderbook
	ctr  int
}

func NewMidSignal() *MidSignal {
	var a MidSignal = MidSignal{}
	print("Mid created\n")
	return &a
}
func (s *MidSignal) initialize(n *Nexus, args map[string]string) {

	switch args["exchange"] {
	case "DYDX":
		s.book = n.DYDX.orderbooks[args["coin"]]
	case "GDAX":
		s.book = n.GDAX.orderbooks[args["coin"]]
	case "KRAKEN":
		s.book = n.KRAKEN.orderbooks[args["coin"]]
	case "BNB":
		s.book = n.BNB.orderbooks[args["coin"]]
	}
}
func (s *MidSignal) forward() float64 {
	// s.ctr += 1
	// return float64(s.ctr)
	return s.book.Mid()
	best_ask := s.book.ask[len(s.book.ask)-2].px
	best_bid := s.book.bid[len(s.book.bid)-2].px

	return (best_ask + best_bid) * 0.5
}
