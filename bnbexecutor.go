package main

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

type BNBexecutor struct {
	PriceTickSize     float64 //https info to find out
	OrderMinSize      float64 //https info to find out
	QuotePrecision    float64
	available_balance float64
	Coin              string
	Orders            []string
	api_key           string
	api_secret        string
	fc                *futures.Client
}

func NewBNBexecutor(coin string) *BNBexecutor {
	d := BNBexecutor{}
	d.Coin = coin
	var testing bool = true
	if testing {
		d.api_key = "245d41e8de1dbc3c18d30a9a7d56ef7a8d11094458a459e298f333bc5eb6b863"
		d.api_secret = "1b23aef395dfb0baf2e03d628d605e745422ab3e958e9175a2b6fbbef27233b7"
		futures.UseTestnet = true

	}
	d.fc = binance.NewFuturesClient(d.api_key, d.api_secret) // USDT-M Futures
	return &d
}

func (b *BNBexecutor) init() {
	res, err := b.fc.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.Assets)
	s, _ := strconv.ParseFloat(res.AvailableBalance, 64)
	b.available_balance = s
	exchangeInfoService := b.fc.NewExchangeInfoService()
	exchangeInfo, err := exchangeInfoService.Do(context.Background())
	exchangeSymbols := exchangeInfo.Symbols
	for _, i := range exchangeSymbols {

		if i.BaseAsset == b.Coin || i.Symbol == b.Coin {
			fmt.Println(i.Symbol, i.BaseAsset, i.MarginAsset)
			b.PriceTickSize = math.Pow(10, -1*float64(i.PricePrecision))
			b.OrderMinSize = math.Pow(10, -1*float64(i.QuantityPrecision))
			b.QuotePrecision = math.Pow(10, -1*float64(i.QuotePrecision))
			break
		}
	}

}
func (b *BNBexecutor) getPosition() float64 {
	a, err := b.fc.NewGetPositionRiskService().Symbol(b.Coin).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	position_obj := a[0]
	position, err := strconv.ParseFloat(position_obj.PositionAmt, 64)
	return position
}

func (b *BNBexecutor) SendALO(Side string, Price float64, Size float64) bool {
	fmt.Println(Price, Size)
	qty := fmt.Sprintf("%.3f", Size)
	px := fmt.Sprintf("%.1f", Price)
	fmt.Println("sending ALO order", "Price = ", px, "Size = ", qty)
	sidevar := futures.SideTypeSell
	if Side == "BUY" {
		sidevar = futures.SideTypeBuy
	}
	order, err := b.fc.NewCreateOrderService().Symbol(b.Coin).Side(sidevar).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).Quantity(qty).
		Price(px).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return order.Status == "NEW"
}

func (b *BNBexecutor) SendIOC(Side string, Price float64, Size float64) bool {
	qty := fmt.Sprintf("%.3f", Size)
	px := fmt.Sprintf("%.1f", Price)
	sidevar := futures.SideTypeSell
	if Side == "BUY" {
		sidevar = futures.SideTypeBuy
	}
	order, err := b.fc.NewCreateOrderService().Symbol(b.Coin).Side(sidevar).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeIOC).Quantity(qty).
		Price(px).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return order.Status == "NEW"
}

func (b *BNBexecutor) SendFOK(Side string, Price float64, Size float64) bool {
	qty := fmt.Sprintf("%.3f", Size)
	px := fmt.Sprintf("%.1f", Price)
	sidevar := futures.SideTypeSell
	if Side == "BUY" {
		sidevar = futures.SideTypeBuy
	}
	order, err := b.fc.NewCreateOrderService().Symbol(b.Coin).Side(sidevar).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeFOK).Quantity(qty).
		Price(px).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return order.Status == "NEW"
}

func (b *BNBexecutor) GetActiveOrders() []RestingOrder {
	fmt.Println("Getting Active Orders")
	a, err := b.fc.NewListOpenOrdersService().Symbol(b.Coin).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	var res []RestingOrder
	for i, _ := range a {
		s := "SELL"
		if a[i].Side == futures.SideTypeBuy {
			s = "BUY"
		}
		ord := RestingOrder{a[i].OrderID, a[i].Price, s, a[i].OrigQuantity, a[i].ExecutedQuantity}
		res = append(res, ord)
	}
	return res
}

func (b *BNBexecutor) CancelOrder(oid int64) bool {
	oid_list := []int64{oid}
	a, err := b.fc.NewCancelMultipleOrdersService().Symbol(b.Coin).OrderIDList(oid_list).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a[0].Status)
	return a[0].Status == "CANCELED"

}

func (b *BNBexecutor) CancelAllOrders() {
	fmt.Println("Cancelling All Orders")
	if len(b.GetActiveOrders()) == 0 {
		return
	}

	err := b.fc.NewCancelAllOpenOrdersService().Symbol(b.Coin).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}
