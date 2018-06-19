package signal
import(
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"fmt"
	"sync"
	"math"
)
func init(){
	replay.SignalGroup = append(replay.SignalGroup, NewSignalAna())
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
	diff := Can.GetMidAverage() - self.can.GetMidAverage()
	if math.Abs(diff) > self.valOut {
		return true
	}
	return false

}
type SignalAna struct{
	//msg
	tag int64
	msgMap map[string]*SignalMsg
	sync.Mutex
}
func NewSignalAna() *SignalAna {
	return &SignalAna{
		msgMap:map[string]*SignalMsg{}}
}
func (self *SignalAna) add(signalmsg *SignalMsg){

	self.Lock()
	defer self.Unlock()

	signalmsg.last = self.msgMap[signalmsg.InsName]
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
	self.msgMap[signalmsg.InsName] = signalmsg

}
func (self *SignalAna) Check(insCache *replay.InstrumentCache){

	var Cache *replay.Cache = nil
	var CacheId int
	for i,ca := range insCache.CacheList {
		if ca.IsUpdate && ca.IsSplit {
			Cache = ca
			CacheId = i
			break
		}
	}
	if Cache == nil {
		return
	}
	if Cache.LastCache == nil {
		return
	}
	//test
	if (math.Abs(Cache.EndJoint.Last.Diff)/math.Abs(Cache.EndJoint.Diff)) < 2{
		return
	}
	HiId := CacheId+1
	if HiId >= len(insCache.CacheList) {
		return
	}

	Cache1 := insCache.CacheList[HiId]
	if Cache1.EndJoint.MaxDiff != 0 {
		return
	}
	if ((Cache.EndJoint.Diff >0) != (Cache1.EndJoint.Diff>0)) {
		return
	}


	/**
	self.lastData.InsCache.CacheList[0].LastCan
	if !da.NowCache.IsSplit {
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

	//   condition 1
	DiffB := math.Abs(JointB.Diff)
	DiffC := math.Abs(JointC.Diff)
	if (DiffC > DiffB) && ((DiffC - DiffB) > da.NowCache.EndJoint.GetLongAve()){
		return
	}
	//

	HightCa := da.InsCache.GetHightCache(da.NowCache)
	if HightCa == nil {
		return
	}
	//   condition 2
	f :=(da.NowCache.EndJoint.Diff >0)
	if f != (HightCa.EndJoint.Diff>0) {
		return
	}
	//

	//   condition 3
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
	//

	self.add(&SignalMsg{
		InsName:da.InsCache.Name,
		timeOut:da.Can.Time - JointB.Cans[0].Time,
		valOut:valOut,
		direction:f,
		can:da.Can})

	**/

}
func (self *SignalAna) Show(){
	fmt.Println("Hello world")
}
