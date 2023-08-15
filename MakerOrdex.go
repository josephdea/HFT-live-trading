package main

import (
	"math"
)

type MakerOrdex struct {
	book             *Orderbook
	mid_bias         float64
	size             float64
	spread_mode      int
	adjust_threshold float64
	spread_scaler    float64
	vol_scaler       float64
	pred_scaler      float64
	Exec             Executor
	position         float64
}

func NewMakerOrdex(nexus *Nexus, Args map[string]string) *MakerOrdex {
	ordex := MakerOrdex{}

	return &ordex
}
func (ordex *MakerOrdex) OnQuery(prediction float64) {
	var spread float64
	if ordex.spread_mode == 0 { //test using spread from effective size
		spread = 0.0001*ordex.book.Mid()*ordex.spread_scaler + 0.4*ordex.book.Spread()*ordex.vol_scaler //pred_scaler * abs_prediction;

	} else if ordex.spread_mode == 1 { //test using bps of the current mid
		spread = 0.0001 * ordex.book.Mid() * ordex.spread_scaler
	}
	newmid := (1+prediction)*ordex.book.Mid() - ordex.mid_bias*ordex.position
	askpx := math.Ceil(newmid + 0.5*spread)
	bidpx := math.Floor(newmid - 0.5*spread)
	askpx = math.Max(ordex.book.BestAsk(), askpx)
	bidpx = math.Min(ordex.book.BestBid(), bidpx)
	asksz := ordex.size
	bidsz := ordex.size
	ordex.position = ordex.Exec.getRecentPosition()
	if asksz != 0 {
		ordex.Exec.SendALO(SIDE.SELL, askpx, asksz)
	}
	if bidsz != 0 {
		ordex.Exec.SendALO(SIDE.BUY, bidpx, bidsz)
	}
	activeSellOrders := ordex.Exec.GetActiveSellOrders()
	activeBuyOrders := ordex.Exec.GetActiveBuyOrders()
	ordersToCancel := make([]string, 0)
	for key, _ := range *activeSellOrders {
		if math.Abs((*activeSellOrders)[key].price-bidpx) > ordex.adjust_threshold { //previous quoted price too far away from the price i want
			ordersToCancel = append(ordersToCancel, key)
		} else { //let the order rest
			bidsz -= (*activeSellOrders)[key].size
		}
	}
	for key, _ := range *activeBuyOrders {
		if math.Abs((*activeBuyOrders)[key].price-bidpx) > ordex.adjust_threshold { //previous quoted price too far away from the price i want
			ordersToCancel = append(ordersToCancel, key)
		} else { //let the order rest
			bidsz -= (*activeBuyOrders)[key].size
		}
	}
	for i := 0; i < len(ordersToCancel); i++ {
		ordex.Exec.CancelOrder(ordersToCancel[i])
	}
	if asksz != 0 {

	}
	if bidsz != 0 {

	}
	if asksz != 0 {
		ordex.Exec.SendALO(SIDE.SELL, askpx, asksz)
	}
	if bidsz != 0 {
		ordex.Exec.SendALO(SIDE.BUY, bidpx, bidsz)
	}
}
