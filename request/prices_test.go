package request
import(
	"testing"
	"fmt"
	"encoding/json"
)
func Test_GetPricing(t *testing.T){
	val:=[]string{"EUR_JPY","EUR_USD"}
	res := GetPricing(val)
	if res != nil {
		for _,p := range res.Prices {
			fmt.Println("ask",p.closeoutAsk,"-")
			for i,a := range p.Asks{
				fmt.Println(i,a)
			}
			fmt.Println("bid",p.closeoutBid,"-")
			for i,a := range p.Bids{
				fmt.Println(i,a)
			}
		}
	}
}
func Test_GetPricingStream(t *testing.T){
	val:=[]string{"EUR_JPY","EUR_USD"}
	var pr Price
	var err error
	GetPricingStream(val,func(da []byte){
		err = json.Unmarshal(da,&pr)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(pr)
	})
}
