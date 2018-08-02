package signal
import(
	"testing"
	"math/rand"
	"fmt"
	"time"
)
func Test_Snap(t *testing.T){

	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	si := NewSignal(nil,nil)
	si.snap = ra.Int63n(1000000)
	snap := ra.Int63n(1000000)
	fmt.Printf("%64b\r\n",si.snap)
	fmt.Printf("%64b\r\n",snap)
	xor:=si.snap ^ snap
	fmt.Printf("%64b\r\n",xor)
	var str string
	for i:=0;i<64;i++ {
		if xor|1 == xor {
			str = fmt.Sprintf("%d %s",i,str)
		}
		xor = xor>>1
	}
	fmt.Println(str)

}
