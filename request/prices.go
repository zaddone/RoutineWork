package request
import(
	"log"
	"net/url"
	"bufio"
	"io"
	//"fmt"
//	"github.com/zaddone/RoutineWork/config"
	"strings"
	//"encoding/json"
)
type Currency string
type HomeConversions struct {
	Currency Currency `json:"currency"`
	AccountGain DecimalNumber `json:"accountGain"`
	AccountLoss DecimalNumber `json:"accountLoss"`
	PositionValue DecimalNumber `json:"positionValue"`
}
type UnitsAvailableDetails struct {
	Long DecimalNumber `json:"long"`
	Short DecimalNumber `json:"short"`
}
type UnitsAvailable struct {
	Default UnitsAvailableDetails `json:"default"`
	ReduceFirst UnitsAvailableDetails `json:"reduceFirst"`
	ReduceOnly UnitsAvailableDetails `json:"reduceOnly"`
	OpenOnly UnitsAvailableDetails `json:"openOnly"`
}
//type DecimalNumber string
type QuoteHomeConversionFactors struct {
	PositiveUnits DecimalNumber `json:"positiveUnits"`
	NegativeUnits DecimalNumber `json:"negativeUnits"`
}
type Price struct {
	Type string `json:"type"`
	Instrument InstrumentName `json:"instrument"`
	Time DateTime `json:"time"`
	Status string `json:"status"`
	Tradeable bool `json:"tradeable"`
	Bids []*PriceBucket `json:"bids"`
	Asks []*PriceBucket `json:"asks"`
	CloseoutBid PriceValue `json:"closeoutBid"`
	CloseoutAsk PriceValue `json:"closeoutAsk"`
	QuoteHomeConversionFactors QuoteHomeConversionFactors `json:"quoteHomeConversionFactors"`
	UnitsAvailable UnitsAvailable `json:"unitsAvailable"`
}
func (self *Price) GetPrice() (pr float64) {

	for _,b := range self.Bids {
		pr+=b.Price.GetPrice()
	}
	for _,b := range self.Asks {
		pr+=b.Price.GetPrice()
	}
	pr+=self.CloseoutAsk.GetPrice()
	pr+=self.CloseoutBid.GetPrice()
	return pr/float64(len(self.Bids)+len(self.Asks)+2)

}
type PricesResponses struct {
	Prices []*Price `json:"prices"`
	HomeConversions []*HomeConversions `json:"HomeConversions"`
	Time DateTime `json:"time"`
}
func GetPricingStream( insNames []string,f func(da []byte) ) error {

	path :=ActiveAccount.GetAccountPathStream()+"/pricing/stream?"+ url.Values{"instruments":[]string{strings.Join(insNames,",")}}.Encode()
	log.Println(path)

	body,err := ClientDOStream(path)
	if err != nil {
		return err
	}
	defer body.Close()
	read := bufio.NewReader(body)
	var by []byte
	for{
		by,err = read.ReadSlice('\n')
		if err == io.EOF {
			break
		}else if err != nil {
			log.Println(err)
			continue
		}
		f(by)
	}
	return nil

}
func GetPricing( insNames []string ) *PricesResponses {

	//da := make(map[string][]*Account)
	var res PricesResponses

	path :=ActiveAccount.GetAccountPath()+"/pricing?"+ url.Values{"instruments":[]string{strings.Join(insNames,",")}}.Encode()
	log.Println(path)

	err := ClientDO(path, &res)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &res

}
