package signal
import(
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"fmt"
	"sync"
	"math"
)
var (

	SignalGroup *SignalAna
)
func init(){
	//replay.SignalSys = append(replay.SignalSys, new(SignalAna))
	SignalGroup = NewSignalAna()

}
type SignalMsg struct {

	InsName string

	timeOut int64
	valOut float64
	direction bool
	//SignalDate
	can *request.Candles

	last *SignalMsg

}
func (self *SignalMsg) check(Can *request.Candles) bool{

	//val  := da.Can.GetMidAverage()
	if Can.Time - self.can.Time >self.timeOut {
		return true
	}
	diff := da.Can.GetMidAverage() - self.can.GetMidAverage()
	if math.Abs(diff) > self.valOut {
		return true
	}
	return false

}
type SignalAna struct{
	//msg
	msgMap map[string]*SignalMsg
	sync.RWMutex
}
func NewSignalAna() *SignalAna {
	return &SignalAna{
		msgMap:map[string]*SignalMsg{}
	}
}
func (self *SignalAna) add(signalmsg *SignalMsg){

	self.RLock()
	signalmsg.last = self.msgMap[signalmsg.InsName]
	self.RUnlock()


	if signalmsg.last != nil {
		if signalmsg.last.direction != signalmsg.direction {
			signalmsg.last = nil
		}else{
			if signalmsg.last.check(signalmsg.can) {
				signalmsg.last = nil
			}else{
				
			}
		}
	}



	self.Lock()
	self.msgMap[signalmsg.InsName] = signalmsg
	self.Unlock()

}
func (self *SignalAna) Check(da *replay.SignalData){

	if !da.Ty {
		return
	}
	if da.NowCache.LastCache == nil {
		return
	}

	JointB := da.NowCache.EndJoint.Last
	if  JointB == nil {
		return
	}
	JointC := JointB.Last
	if  JointC == nil {
		return
	}
	DiffB := math.Abs(JointB.Diff)
	DiffC := math.Abs(JointC.Diff)
	if (DiffC > DiffB) && ((DiffC - DiffB) > da.NowCache.EndJoint.GetLongAve()){
		return
	}
	//da.Can.Time - JointB.Cans[0].Time

	HightCa := da.InsCache.GetHightCache(da.NowCache)
	if HightCa == nil {
		return
	}
	f :=(da.NowCache.EndJoint.Diff >0)
	if f != (HightCa.EndJoint.Diff>0) {
		return
	}
	le :=len(HightCa.EndJoint.Cans)
	if  le < 3 {
		return
	}
	absDiff := math.Abs(HightCa.EndJoint.Diff)
	ave := HightCa.EndJoint.GetLongAve()

	if absDiff < ave {
		return
	}
	if HigitCa.EndJoint.Last == nil {
		return
	}
	valOut := math.Abs(HigitCa.EndJoint.Last.Diff) - absDiff
	if valOut < absDiff {
		return
	}

	self.add(&SignalMsg{
		InsName:da.InsCache.Name,
		timeOut:da.Can.Time - JointB.Cans[0].Time,
		valOut:valOut,
		direction:f,
		can:da.Can})


}
func (self *SignalAna) Show(){
	fmt.Println("Hello world")
}
