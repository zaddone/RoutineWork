package signal
import(
	"github.com/zaddone/RoutineWork/replay"
	"sync"
	"fmt"
)
type Ana struct{

	msgMap map[string]*Signal
	sync.Mutex

}
func NewAna() *Ana {
	return &Ana{
		msgMap:map[string]*Signal{}}
}
func (self *Ana) add(msg *Msg,Ins *replay.InstrumentCache){

	key := Ins.Ins.Name
	sig := self.msgMap[key]
	if sig == nil {
		self.Lock()
		self.msgMap[key] = &Signal{
			Msg:msg,
			InsCache:Ins}
		self.Unlock()
		return
	}
	//if !sig.Msg.merge(msg) {
	//	sig.Msg.Close(sig.InsCache.GetBaseCan().GetMidAverage())
	//	sig.Msg = msg
	//}
	sig.Add(msg)
	//if sig.Msg.Count()>1{
	//	sig.Msg.Show(0)
	//}
	//sig.Msg.Show(0)
	sig.PostOrder(sig.Msg)

}

func (self *Ana) Check(InsCache *replay.InstrumentCache){

	func(){
		sig := self.msgMap[InsCache.Ins.Name]
		if sig == nil || sig.Msg == nil {
			return
		}

		baseCache := InsCache.GetBaseCache()
		sig.Msg = sig.Msg.readAll(func(mg *Msg) (out bool){
			return mg.testOut(baseCache)
		})

		//if sig.Msg == nil {
		//	return
		//}
		//sig.Msg = sig.Msg.readAll(func(mg *Msg) (out bool){
		//	if !mg.endCache.IsSplit {
		//		return false
		//	}
		//	return mg.testClose()

		//})
	}()
	f := func(Cache *replay.Cache,CacheId int) bool {
		dif := (Cache.EndJoint.Diff >0)
		end := InsCache.Selective(dif,CacheId)

		var sl,_sl float64
		var endId int
		//Max:=Cache.GetInsCache().Ins.MinimumTrailingStopDistance *1.5
		for i:= CacheId; i<end; i++ {
			_sl = InsCache.CacheList[i].EndJoint.MaxDiff
			//if _sl > sl && _sl < Max {
			if _sl > sl {
				sl = _sl
				endId = i
			}
		}
		if sl == 0 {
			return false
		}
		if dif {
			sl = -sl
		}

		endCache := InsCache.CacheList[endId]

		self.add(&Msg{
			tp:-sl,
			sl:sl,
			post:false,
			//scale: Cache.Scale,
			endJoint:endCache.EndJoint,
			endCache:endCache,
			joint:Cache.GetMinCache().EndJoint,
			can:Cache.GetMinCan()},
			InsCache)
		return true

	}
	for i,tmpca := range InsCache.CacheList[:len(InsCache.CacheList)-3] {
		if tmpca.IsSplit {
			if f(tmpca,i) {
				break
			}
		}
	}

}

func (self *Ana) Show(){
	fmt.Println("Hello world")
}
