package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	//"math"
)

type PriceValue string
type DateTime string
type AccountID string
type RequestID string
type OrderID string
type InstrumentName string
type DecimalNumber string
type AccountUnits string
type TransactionType string
type TransactionID string

type PriceBucket struct {
	Price     PriceValue `json:"price"`
	Liquidity interface{} `json:"liquidity"`
}
func (self PriceValue) GetPrice() float64 {
	v,err := strconv.ParseFloat(string(self),64)
	if err != nil {
		log.Fatalln(err)
		return 0
	}
	return v
}
type ClientPrice struct {
	Bids        []PriceBucket `json:"bids"`
	Asks        []PriceBucket `json:"Asks"`
	CloseoutBid PriceValue    `json:"closeoutBid"`
	CloseoutAsk PriceValue    `json:"closeoutAsk"`
	Timestamp   DateTime      `json:"timestamp"`
}
type TradeO struct {
	TradeID string `json:"tradeID"`
	Units   string `json:"units"`
}

type TradeReduce struct {
	TradeID    string `json:"tradeID"`
	Units      string `json:"units"`
	RealizedPL string `json:"realizedPL"`
	Financing  string `json:"financing"`
}
type Transaction struct {
	Id        TransactionID `json:"id"`
	Time      DateTime      `json:"time"`
	UserID    int           `json:"userID"`
	AccountID AccountID     `json:"accountID"`
	BatchID   TransactionID `json:"batchID"`
	RequestID RequestID     `json:"requestID"`
}
type OrderCancelT struct {
	Transaction
	Type              TransactionType `json:"type"`
	OrderID           OrderID         `json:"orderID"`
	ClientOrderID     OrderID         `json:"clientOrderID"`
	Reason            string          `json:"reason"`
	ReplacedByOrderID OrderID         `json:"replacedByOrderID"`
}
type OrderFillT struct {
	Transaction
	Type           TransactionType `json:"type"`
	OrderID        OrderID         `json:"orderID"`
	ClientOrderID  OrderID         `json:"clientOrderID"`
	Instrument     InstrumentName  `json:"instrument"`
	Units          DecimalNumber   `json:"units"`
	Price          PriceValue      `json:"price"`
	FullPrice      ClientPrice     `json:"fullPrice"`
	Reason         string          `json:"reason"`
	Pl             AccountUnits    `json:"pl"`
	Financing      AccountUnits    `json:"financing"`
	Commission     AccountUnits    `json:"commission"`
	AccountBalance AccountUnits    `json:"accountBalance"`
	TradeOpened    TradeO          `json:"tradeOpened"`
	TradesClosed   []TradeReduce   `json:"tradeOpened"`
	TradeReduced   TradeReduce     `json:"tradeReduced"`
}
type OrderResponse struct {

	OrderRejectTransaction Transaction     `json:"orderRejectTransaction"`
	RelatedTransactionIDs  []TransactionID `json:"relatedTransactionIDs"`

	OrderCreateTransaction  Transaction  `json:"orderCreateTransaction"`
	OrderFillTransaction    OrderFillT   `json:"orderFillTransaction"`
	OrderCancelTransaction  OrderCancelT `json:"orderCancelTransaction"`
	OrderReissueTransaction Transaction  `json:"orderReissueTransaction"`

	LastTransactionID TransactionID `json:"lastTransactionID"`
	ErrorCode         string        `json:"errorCode"`
	ErrorMessage      string        `json:"errorMessage"`

}

type ClientExt struct {
	Id      string `json:"id,omitempty"`
	Tag     string `json:"tag,omitempty"`
	Comment string `json:"comment,omitempty"`
}
type Details struct {
	Price            string     `json:"price,omitempty"`
	TimeInForce      string     `json:"timeInForce,omitempty"`
	GtdTime          string     `json:"gtdTime,omitempty"`
	ClientExtensions *ClientExt `json:"clientExtensions,omitempty"`
}
type TrailingStopLossDetails struct {
	Distance         string     `json:"distance,omitempty"`
	TimeInForce      string     `json:"timeInForce,omitempty"`
	GtdTime          string     `json:"gtdTime,omitempty"`
	ClientExtensions *ClientExt `json:"clientExtensions,omitempty"`
}
type MarketOrderRequest struct {
	Type                   string                   `json:"type,omitempty"`
	Instrument             string                   `json:"instrument,omitempty"`
	Units                  string                   `json:"units,omitempty"`
	TimeInForce            string                   `json:"timeInForce,omitempty"`
	PriceBount             string                   `json:"priceBount,omitempty"`
	PositionFill           string                   `json:"positionFill,omitempty"`
	ClientExtensions       *ClientExt               `json:"clientExtensions,omitempty"`
	TakeProfitOnFill       *Details                 `json:"takeProfitOnFill,omitempty"`
	StopLossOnFill         *Details                 `json:"stopLossOnFill,omitempty"`
	TrailingStopLossOnFill *TrailingStopLossDetails `json:"trailingStopLossOnFill,omitempty"`
	TradeClientExtensions  *ClientExt               `json:"tradeClientExtensions,omitempty"`
}
func NewMarketOrderRequest(InsName string) *MarketOrderRequest {
	return &MarketOrderRequest{
		Type:"MARKET",
		Instrument : InsName,
		TimeInForce : "FOK",
		PositionFill : "DEFAULT",
		PriceBount : "2"}
}

func (self *MarketOrderRequest) Init(InsName string) {
	self.Type = "MARKET"
	self.Instrument = InsName
	//	self.Units = "100"
	//	self.Units = fmt.Sprintf("%d",int(math.Pow(10,Instr.DisplayPrecision)*Instr.MinimumTradeSize))
	self.TimeInForce = "FOK"
	self.PositionFill = "DEFAULT"
	self.PriceBount = "2"
}
func (self *MarketOrderRequest) SetStopLossDetails(price string) {
	self.StopLossOnFill = &Details{
		Price : price	}
}
func (self *MarketOrderRequest) SetTakeProfitDetails(price string) {
	self.TakeProfitOnFill = &Details{
		Price : price	}
}

func (self *MarketOrderRequest) SetTrailingStopLossDetails(dif string) {
	self.TrailingStopLossOnFill = & TrailingStopLossDetails{
		Distance : dif	}
}

func (self *MarketOrderRequest) SetUnits(units int) {
	self.Units = fmt.Sprintf("%d", units)
}

func HandleOrder(InsName string,unit int, dif , Tp, Sl string) (mr OrderResponse, err error) {

	path := ActiveAccount.GetAccountPath()
	path += "/orders"
	//Val := make(map[string]*MarketOrderRequest)
	order := NewMarketOrderRequest(InsName)
	//order.Init()
	order.SetUnits(unit)
	if dif != "" {
		order.SetTrailingStopLossDetails(dif)
	}
	if Sl != "" {
		order.SetStopLossDetails(Sl)
	}

	if Tp != "" {
		order.SetTakeProfitDetails(Tp)
	}
	//Val["order"] = order

	da, err := json.Marshal(map[string]*MarketOrderRequest{"order":order})
	if err != nil {
		panic(err)
	}
	err = ClientPost(path, bytes.NewReader(da), &mr)
	return mr, err

}

