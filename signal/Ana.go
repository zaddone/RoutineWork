package signal
import(
	//"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/config"
	"sync"
	"fmt"
	"math"
)
type Ana struct{

	msgMap map[string]*Signal
	sync.RWMutex
	//sync.Mutex

}
func NewAna() *Ana {
	return &Ana{msgMap:map[string]*Signal{}}
}
func (self *Ana) Get(k string) (s *Signal) {
	self.RLock()
	s = self.msgMap[k]
	self.RUnlock()
	return
}
func (self *Ana) Set(k string,s *Signal) {
	self.Lock()
	self.msgMap[k] = s
	self.Unlock()
}
func (self *Ana) Update(ca config.Cache) {

	key := ca.GetInsCache().GetInsName()
	sig := self.Get(key)
	if sig == nil {
		sig = NewSignal(ca.GetInsCache())
		self.Set(key,sig)
	}
	sig.UpdateJoint(ca)

}

func (self *Ana) Check (InsCache config.Instrument){

	sig := self.msgMap[InsCache.GetInsName()]
	if sig == nil {
		return
	}
	ids := sig.CheckSnap()
	if len(ids)  == 0 {
		return
	}
	cfunc := func(id int) bool {
		big := id+1
		if big == InsCache.GetCacheLen() {
			return false
		}

		cache := InsCache.GetCache(id)
		jc := sig.Joint.Get(cache.GetScale())

		var key uint64
		f :=jc.EndJoint.Diff>0
		for i:=big;i<InsCache.GetCacheLen();i++{
			_jc := sig.Joint.Get(InsCache.GetCache(i).GetScale())
			if _jc == nil {
				return false
			}
			if (_jc.EndJoint.Diff >0 ) == f {
				key ++
			}else if key == 0 {
				return false
			}
			key = key << 1
		}
		if jc.EndJoint.Diff > 0 {
			key++
		}
		//fmt.Printf("%16b--------\r\n",key)
		start:= true
		dis := sig.InsCache.GetInsStopDis()
		minDis := dis*1.5
		maxDis := dis*3.0
		diff := math.Abs(jc.EndJoint.Diff)
		if  diff < minDis || diff>maxDis {
			start = false
		}

		jc.AddTmpSnap(NewSnap(diff,jc.EndJoint,cache.GetlastCan(),start,key))
		return true

	}
	//id:=ids[0]
	for _,id := range ids {
		if cfunc(id) {
			//sig.ShowMergeRatio()
			return
		}

	}

}


func (self *Ana) Show(){
	fmt.Println("Hello world")
}
