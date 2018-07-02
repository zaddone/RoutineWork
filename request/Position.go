package request

import (
	"bytes"
	"encoding/json"
	//	"fmt"
	"strconv"
)

//type TradeReduce struct {
//	TradeID    TradeID       `json:"tradeID"`
//	Units      DecimalNumber `json:"units"`
//	RealizedPL AccountUnits  `json:"RealizedPL"`
//	Financing  AccountUnits  `json:"financing"`
//}
type TradeID string
type TradeOpen struct {
	TradeID          TradeID       `json:"tradeID"`
	Units            DecimalNumber `json:"units"`
	ClientExtensions ClientExt     `json:"clientExtensions"`
}

type MarketOrderDelayedTradeClose struct {
	TradeID             TradeID       `json:"tradeID"`
	ClientTradeID       TradeID       `json:"clientTradeID"`
	SourceTransactionID TransactionID `json:"sourceTransactionID"`
}

type MarketOrderMarginCloseout struct {
	Reason string `json:"reason"`
}

type MarketOrderPositionCloseout struct {
	Instrument InstrumentName `json:"instrument"`
	Units      string         `json:"units"`
}
type MarketOrderTradeClose struct {
	TradeID       TradeID `json:"tradeID"`
	ClientTradeID string  `json:"clientTradeID"`
	Units         string  `json:"units"`
}

type MarketOrderTransaction struct {
	Id                     TransactionID                `json:"id"`
	Time                   DateTime                     `json:"time"`
	UserID                 int                          `json:"userID"`
	AccountID              AccountID                    `json:"accountID"`
	BatchID                TransactionID                `json:"batchID"`
	RequestID              RequestID                    `json:"requestID"`
	Type                   string                       `json:"type"`
	Instrument             InstrumentName               `json:"Instrument"`
	Units                  DecimalNumber                `json:"units"`
	TimeInForce            string                       `json:"timeInForce"`
	PriceBound             PriceValue                   `json:"priceBound"`
	PositionFill           string                       `json:"positionFill"`
	TradeClose             MarketOrderTradeClose        `json:"tradeClose"`
	LongPositionCloseout   MarketOrderPositionCloseout  `json:"longPositionCloseout"`
	ShortPositionCloseout  MarketOrderPositionCloseout  `json:"shortPositionCloseout"`
	MarginCloseout         MarketOrderMarginCloseout    `json:"marginCloseout"`
	DelayedTradeClose      MarketOrderDelayedTradeClose `json:"delayedTradeClose"`
	Reason                 string                       `json:"reason"`
	ClientExtensions       ClientExt                    `json:"clientExtensions"`
	TakeProfitOnFill       Details                      `json:"takeProfitOnFill"`
	StopLossOnFill         Details                      `json:"stopLossOnFill"`
	TrailingStopLossOnFill TrailingStopLossDetails      `json:"trailingStopLossOnFill"`
	TradeClientExtensions  ClientExt                    `json:"tradeClientExtensions"`
}
type OrderCancelTransaction struct {
	Id                TransactionID `json:"id"`
	Time              DateTime      `json:"time"`
	UserID            int           `json:"userID"`
	AccountID         AccountID     `json:"accountID"`
	BatchID           TransactionID `json:"batchID"`
	RequestID         RequestID     `json:"requestID"`
	Type              string        `json:"type"`
	OrderID           OrderID       `json:"orderID"`
	ClientOrderID     OrderID       `json:"clientOrderID"`
	Reason            string        `json:"reason"`
	ReplacedByOrderID OrderID       `json:"replacedByOrderID"`
}
type OrderFillTransaction struct {
	Id             TransactionID  `json:"id"`
	Time           DateTime       `json:"time"`
	UserID         int            `json:"userID"`
	AccountID      AccountID      `json:"accountID"`
	BatchID        TransactionID  `json:"batchID"`
	RequestID      RequestID      `json:"requestID"`
	Type           string         `json:"type"`
	OrderID        OrderID        `json:"orderID"`
	ClientOrderID  OrderID        `json:"clientOrderID"`
	Instrument     InstrumentName `json:"Instrument"`
	Units          DecimalNumber  `json:"units"`
	Price          PriceValue     `json:"price"`
	FullPrice      ClientPrice    `json:"fullPrice"`
	Reason         string         `json:"reason"`
	Pl             AccountUnits   `json:"pl"`
	Financing      AccountUnits   `json:"financing"`
	Commission     AccountUnits   `json:"commission"`
	AccountBalance AccountUnits   `json:"accountBalance"`
	TradeOpened    TradeOpen      `json:"tradeOpened"`
	TradesClosed   []TradeReduce  `json:"tradesClosed"`
	TradeReduced   TradeReduce    `json:"tradeReduced"`
}
type PositionResponses struct {
	LongOrderCreateTransaction MarketOrderTransaction `json:"longOrderCreateTransaction"`
	LongOrderFillTransaction   OrderFillTransaction   `json:"LongOrderFillTransaction"`
	LongOrderCancelTransaction OrderCancelTransaction `json:"LongOrderCancelTransaction"`

	ShortOrderCreateTransaction MarketOrderTransaction `json:"shortOrderCreateTransaction"`
	ShortOrderFillTransaction   OrderFillTransaction   `json:"shortOrderFillTransaction"`
	ShortOrderCancelTransaction OrderCancelTransaction `json:"shortOrderCancelTransaction"`

	RelatedTransactionIDs []TransactionID `json:"relatedTransactionIDs"`
	LastTransactionID     TransactionID   `json:"lastTransactionID"`
}

func ClosePosition(InsName string,longUnits string) (tr float64, err error) {

	path := ActiveAccount.GetAccountPath()
	path += "/positions/" + InsName + "/close"

	val := make(map[string]string)
	//val["longUnits"] = "ALL"
	val["longUnits"] = longUnits
	da, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	var mr PositionResponses
	err = ClientPut(path, bytes.NewReader(da), &mr)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(string(mr.LongOrderFillTransaction.Pl),64)
	//fmt.Println(mr.LongOrderFillTransaction.Pl)
	//le := len(mr.LongOrderFillTransaction.TradesClosed)
	//if le > 0 {
	//	return strconv.ParseFloat(mr.LongOrderFillTransaction.TradesClosed[le-1].RealizedPL, 64)
	//}
	//return 0, nil
}
