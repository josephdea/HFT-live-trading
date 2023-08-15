package main

import (
	"crypto/hmac"
	"crypto/sha256"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/huandu/xstrings"

	//"json"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	phash "github.com/consensys/gnark-crypto/ecc/stark-curve/pedersen-hash"
	gojson "github.com/goccy/go-json"
	"github.com/shopspring/decimal"
	//"github.com/shopspring/decimal"
	//"github.com/ericlagergren/decimal"
)

type PositionStruct struct {
	Market     string `json:"market"`
	Status     string `json:"status"`
	Side       string `json:"side"`
	Size       string `json:"size"`
	Maxsize    string `json:"maxSize"`
	EntryPrice string `json:"entryPrice"`
}
type Account struct {
	PositionID         string                    `json:"positionId"`
	Equity             string                    `json:"equity"`
	FreeCollat         string                    `json:"freCollateral"`
	PendingDeposits    string                    `json:"pendingDeposits"`
	PendingWithdrawals string                    `json:"pendingWithdrawals"`
	OpenPositions      map[string]PositionStruct `json:"openPositions"`
	AccountNumber      string                    `json:"accountNumber"`
	Id                 string                    `json:"id"`
	QuoteBalance       string                    `json:"quoteBalance"`
	CreatedAt          string                    `json:"createdAt"`
}
type Accounts struct {
	Accs []Account `json:"accounts"`
}

func IntToHex32(x *big.Int) string {
	str := x.Text(16)
	return xstrings.RightJustify(str, 64, "0")
}
func SerializeSignature(r, s *big.Int) string {
	return IntToHex32(r) + IntToHex32(s)
}
func pedersen_hash(a string, b string) string {
	// fmt.Println(a, b)
	a_fp, _ := new(fp.Element).SetString(a)
	b_fp, _ := new(fp.Element).SetString(b)
	ans := phash.Pedersen(a_fp, b_fp)
	return ans.Text(10)
}

type OrderSignParam struct {
	NetworkId  int    `json:"network_id"` // 1 MAINNET 3 ROPSTEN
	PositionId int64  `json:"position_id"`
	Market     string `json:"market"`
	Side       string `json:"side"`
	HumanSize  string `json:"human_size"`
	HumanPrice string `json:"human_price"`
	LimitFee   string `json:"limit_fee"`
	ClientId   string `json:"clientId"`
	Expiration string `json:"expiration"` // 2006-01-02T15:04:05.000Z
}
type OrderSigner struct {
	param OrderSignParam
	msg   struct {
		OrderType               string   `json:"order_type"`
		AssetIdSynthetic        *big.Int `json:"asset_id_synthetic"`
		AssetIdCollateral       *big.Int `json:"asset_id_collateral"`
		AssetIdFee              *big.Int `json:"asset_id_fee"`
		QuantumAmountSynthetic  *big.Int `json:"quantum_amount_synthetic"`
		QuantumAmountCollateral *big.Int `json:"quantum_amount_collateral"`
		QuantumAmountFee        *big.Int `json:"quantum_amount_fee"`
		IsBuyingSynthetic       bool     `json:"is_buying_synthetic"`
		PositionId              *big.Int `json:"position_id"`
		Nonce                   *big.Int `json:"nonce"`
		ExpirationEpochHours    *big.Int `json:"expiration_epoch_hours"`
	}
	currency        string
	assetIdSyn      *big.Int
	assetId         *big.Int
	resolutionC     decimal.Decimal
	limitFeeRounded decimal.Decimal
}

func NonceByClientId(clientId string) *big.Int {
	h := sha256.New()
	h.Write([]byte(clientId))

	a := new(big.Int)
	a.SetBytes(h.Sum(nil))
	res := a.Mod(a, big.NewInt(NONCE_UPPER_BOUND_EXCLUSIVE))
	return res
}
func (s *OrderSigner) initMsg() error {
	currency := strings.Split(s.param.Market, "-")[0] // EOS-USD -> EOS
	s.currency = currency
	assetIdSyn, ok := big.NewInt(0).SetString(SYNTHETIC_ID_MAP[currency], 0) // with prefix: 0x
	s.assetIdSyn = assetIdSyn
	if !ok {
		return errors.New("invalid market: " + s.param.Market)
	}
	assetId := COLLATERAL_ASSET_ID_BY_NETWORK_ID[s.param.NetworkId] // asset id
	s.assetId = assetId
	if assetId == nil {
		return errors.New(fmt.Sprintf("invalid network_id: %v", s.param.NetworkId))
	}
	exp, err := time.Parse("2006-01-02T15:04:05.000Z", s.param.Expiration)
	if err != nil {
		return err
	}
	resolutionC := decimal.NewFromInt(ASSET_RESOLUTION[currency])
	s.resolutionC = resolutionC
	price, err := decimal.NewFromString(s.param.HumanPrice)
	if err != nil {
		return err
	}
	size, err := decimal.NewFromString(s.param.HumanSize)
	if err != nil {
		return err
	}
	var quantumsAmountSynthetic = decimal.NewFromFloat(0)
	isBuy := s.param.Side == "BUY"
	if isBuy {
		quantumsAmountSynthetic = size.Mul(price).Mul(resolutionUsdc).RoundUp(0)
	} else {
		quantumsAmountSynthetic = size.Mul(price).Mul(resolutionUsdc).RoundDown(0)
	}
	limitFeeRounded, err := decimal.NewFromString(s.param.LimitFee)
	s.limitFeeRounded = limitFeeRounded
	if err != nil {
		return err
	}
	s.msg.OrderType = "LIMIT_ORDER_WITH_FEES"
	s.msg.AssetIdSynthetic = assetIdSyn
	s.msg.AssetIdCollateral = assetId
	s.msg.AssetIdFee = assetId
	s.msg.QuantumAmountSynthetic = size.Mul(resolutionC).BigInt()
	s.msg.QuantumAmountCollateral = quantumsAmountSynthetic.BigInt()
	s.msg.QuantumAmountFee = limitFeeRounded.Mul(quantumsAmountSynthetic).RoundUp(0).BigInt()
	s.msg.IsBuyingSynthetic = isBuy
	s.msg.PositionId = big.NewInt(s.param.PositionId)
	s.msg.Nonce = NonceByClientId(s.param.ClientId)
	s.msg.ExpirationEpochHours = big.NewInt(int64(math.Ceil(float64(exp.Unix())/float64(ONE_HOUR_IN_SECONDS))) + ORDER_SIGNATURE_EXPIRATION_BUFFER_HOURS)
	//s.msg.ExpirationEpochHours = big.NewInt(int64(466342))
	return nil
}

// what changes and needs to be updated?
// side, price, size, clientid,expiration
func (s *OrderSigner) fastUpdate(Side int, Price string, Size string) {
	// if Side == SIDE.BUY {
	// 	s.param.Side = "BUY"
	// } else {
	// 	s.param.Side = "SELL"
	// }
	price, err := decimal.NewFromString(Price)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	size, err := decimal.NewFromString(Size)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exp, err := time.Parse("2006-01-02T15:04:05.000Z", time.Now().Add(time.Hour).UTC().Format("2006-01-02T15:04:05.000Z"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var quantumsAmountSynthetic = decimal.NewFromFloat(0)
	isBuy := Side < 0
	if isBuy {
		quantumsAmountSynthetic = size.Mul(price).Mul(resolutionUsdc).RoundUp(0)
	} else {
		quantumsAmountSynthetic = size.Mul(price).Mul(resolutionUsdc).RoundDown(0)
	}
	// s.msg.OrderType = "LIMIT_ORDER_WITH_FEES"
	// s.msg.AssetIdSynthetic = assetIdSyn
	// s.msg.AssetIdCollateral = assetId
	// s.msg.AssetIdFee = assetId
	s.msg.QuantumAmountSynthetic = size.Mul(s.resolutionC).BigInt()
	s.msg.QuantumAmountCollateral = quantumsAmountSynthetic.BigInt()
	s.msg.QuantumAmountFee = s.limitFeeRounded.Mul(quantumsAmountSynthetic).RoundUp(0).BigInt()
	s.msg.IsBuyingSynthetic = isBuy
	s.msg.Nonce = NonceByClientId(random_client_id())
	s.msg.ExpirationEpochHours = big.NewInt(int64(math.Ceil(float64(exp.Unix())/float64(ONE_HOUR_IN_SECONDS))) + ORDER_SIGNATURE_EXPIRATION_BUFFER_HOURS)
}

type PedersenCfg struct {
	Comment        string        `json:"_comment"`
	FieldPrime     *big.Int      `json:"FIELD_PRIME"`
	FieldGen       int           `json:"FIELD_GEN"`
	EcOrder        *big.Int      `json:"EC_ORDER"`
	ALPHA          int           `json:"ALPHA"`
	BETA           *big.Int      `json:"BETA"`
	ConstantPoints [][2]*big.Int `json:"CONSTANT_POINTS"`
}

var pedersenCfg PedersenCfg

var EC_ORDER = new(big.Int)
var FIELD_PRIME = new(big.Int)

func init() {
	_ = gojson.Unmarshal([]byte(pedersenParams), &pedersenCfg)
	EC_ORDER = pedersenCfg.EcOrder
	FIELD_PRIME = pedersenCfg.FieldPrime
}

func PedersenHash(str ...string) string {
	NElementBitsHash := FIELD_PRIME.BitLen()
	point := pedersenCfg.ConstantPoints[0]
	wg := sync.WaitGroup{}
	for i, s := range str {
		go func(i int, s string) {
			wg.Add(1)
			x, _ := big.NewInt(0).SetString(s, 10)
			pointList := pedersenCfg.ConstantPoints[2+i*NElementBitsHash : 2+(i+1)*NElementBitsHash]
			n := big.NewInt(0)
			for _, pt := range pointList {
				n.And(x, big.NewInt(1))
				if n.Cmp(big.NewInt(0)) > 0 {
					point = eccAdd(point, pt, FIELD_PRIME)
				}
				x = x.Rsh(x, 1)
			}
			wg.Done()
		}(i, s)
	}
	wg.Wait()
	return point[0].String()
}
func getHash(str1, str2 string) string {
	return PedersenHash(str1, str2)
}
func (s *OrderSigner) getHash() (string, error) {
	var assetIdSell, assetIdBuy, quantumsAmountSell, quantumsAmountBuy *big.Int
	if s.msg.IsBuyingSynthetic {
		assetIdSell = s.msg.AssetIdCollateral
		assetIdBuy = s.msg.AssetIdSynthetic
		quantumsAmountSell = s.msg.QuantumAmountCollateral
		quantumsAmountBuy = s.msg.QuantumAmountSynthetic
	} else {
		assetIdSell = s.msg.AssetIdSynthetic
		assetIdBuy = s.msg.AssetIdCollateral
		quantumsAmountSell = s.msg.QuantumAmountSynthetic
		quantumsAmountBuy = s.msg.QuantumAmountCollateral
	}
	fee := s.msg.QuantumAmountFee
	nonce := s.msg.Nonce
	// part1
	part1 := big.NewInt(0).Set(quantumsAmountSell)
	//fmt.Println("part 1", part1)
	part1.Lsh(part1, ORDER_FIELD_BIT_LENGTHS["quantums_amount"])
	part1.Add(part1, quantumsAmountBuy)
	part1.Lsh(part1, ORDER_FIELD_BIT_LENGTHS["quantums_amount"])
	part1.Add(part1, fee)
	part1.Lsh(part1, ORDER_FIELD_BIT_LENGTHS["nonce"])
	part1.Add(part1, nonce)
	//fmt.Println("part 1", part1)
	// part2
	part2 := big.NewInt(ORDER_PREFIX)
	for i := 0; i < 3; i++ {
		part2.Lsh(part2, ORDER_FIELD_BIT_LENGTHS["position_id"])
		part2.Add(part2, s.msg.PositionId)
	}
	part2.Lsh(part2, ORDER_FIELD_BIT_LENGTHS["expiration_epoch_hours"])
	part2.Add(part2, s.msg.ExpirationEpochHours)
	part2.Lsh(part2, ORDER_PADDING_BITS)
	//fmt.Println("part 2", part2)
	// pedersen hash
	assetHash := pedersen_hash(pedersen_hash(assetIdSell.String(), assetIdBuy.String()), s.msg.AssetIdFee.String())
	// fmt.Println(assetIdSell.String(), assetIdBuy.String())
	// fmt.Println("int hash", getHash(assetIdSell.String(), assetIdBuy.String()))
	// fmt.Println("int1 hash", pedersen_hash(assetIdSell.String(), assetIdBuy.String()))
	// fmt.Println("assetHash", assetHash)
	part1Hash := pedersen_hash(assetHash, part1.String())
	part2Hash := pedersen_hash(part1Hash, part2.String())
	return part2Hash, nil
}

func random_client_id() string {
	rand.Seed(time.Now().UnixNano())
	res, _ := strconv.Atoi(fmt.Sprintf("%f", rand.Float64())[2:])
	return strconv.Itoa(res)

}

func generateSignature(msg string, secret string) string {
	res, _ := b64.RawURLEncoding.DecodeString(secret)
	h := hmac.New(sha256.New, res)
	h.Write([]byte(msg))
	//fmt.Println("debug", h.Sum(nil))
	sha := b64.URLEncoding.EncodeToString(h.Sum(nil))
	return sha
}
func (d DYDXexecutor) SignRequest(req *http.Request, method string, requestPath string) {
	curTime := d.GetISOTime()
	msg := curTime + method + requestPath
	HMAC := generateSignature(msg, d.ApiKey["secret"])
	req.Header.Set("dydx-signature", HMAC)
	req.Header.Set("DYDX-API-KEY", d.ApiKey["public_key"])
	req.Header.Set("DYDX-TIMESTAMP", curTime)
	req.Header.Set("DYDX-PASSPHRASE", d.ApiKey["passphrase"])

}
func (d DYDXexecutor) SignPostRequest(req *http.Request, method string, requestPath string, data map[string]string) {
	curTime := d.GetISOTime()
	jsonStr, _ := gojson.Marshal(data)
	//fmt.Println(string(jsonStr))
	msg := curTime + method + requestPath + string(jsonStr)
	HMAC := generateSignature(msg, d.ApiKey["secret"])
	req.Header.Set("dydx-signature", HMAC)
	req.Header.Set("DYDX-API-KEY", d.ApiKey["public_key"])
	req.Header.Set("DYDX-TIMESTAMP", curTime)
	req.Header.Set("DYDX-PASSPHRASE", d.ApiKey["passphrase"])
}

func (d DYDXexecutorv2) SignRequest(req *http.Request, method string, requestPath string) {
	curTime := d.GetISOTime()
	msg := curTime + method + requestPath
	HMAC := generateSignature(msg, d.ApiKey["secret"])
	req.Header.Set("dydx-signature", HMAC)
	req.Header.Set("DYDX-API-KEY", d.ApiKey["public_key"])
	req.Header.Set("DYDX-TIMESTAMP", curTime)
	req.Header.Set("DYDX-PASSPHRASE", d.ApiKey["passphrase"])

}
func (d DYDXexecutorv2) SignPostRequest(req *http.Request, method string, requestPath string, data map[string]string) {
	curTime := d.GetISOTime()
	jsonStr, _ := gojson.Marshal(data)
	//fmt.Println(string(jsonStr))
	msg := curTime + method + requestPath + string(jsonStr)
	HMAC := generateSignature(msg, d.ApiKey["secret"])
	req.Header.Set("dydx-signature", HMAC)
	req.Header.Set("DYDX-API-KEY", d.ApiKey["public_key"])
	req.Header.Set("DYDX-TIMESTAMP", curTime)
	req.Header.Set("DYDX-PASSPHRASE", d.ApiKey["passphrase"])
}
