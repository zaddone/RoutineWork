package signal
import(
	"github.com/zaddone/RoutineWork/config"
	"math"
	//"fmt"
	"sync"
)
type JointMap struct {
	sync.RWMutex
	jointMap map[int64] *JointCache
}
func NewJointMap () *JointMap {
	return &JointMap{jointMap:map[int64]*JointCache{}}
}
func (self *JointMap) Set(k int64,v *JointCache) {
	self.Lock()
	self.jointMap[k] = v
	self.Unlock()
}
func (self *JointMap) Get(k int64) (j *JointCache){
	self.RLock()
	j = self.jointMap[k]
	self.RUnlock()
	return
}
type snap struct {

	diff float64
	jo *Joint
	ca config.Candles
	direction bool
	start bool
	forecast bool
	Post bool
	//key uint64
	_key []uint64

}

func NewSnap(diff float64,jo *Joint,ca config.Candles,start bool,key uint64) (sn *snap) {
	sn = &snap{diff:diff,
		jo:jo,
		ca:ca,
		start:start,
		//key:key,
		direction:jo.Diff>0}
	//sn.setKeyList(key)
	return sn

}

func (self *snap) setKeyList(key uint64){

	//key := self.key
	var d uint64
	if self.direction {
		d = 1
	}
	for i:=0;i<64;i++{
		key =  key>>1
		if key == 0 {
			break
		}
		if key|1 != key {
			continue
		}
		self._key =append(self._key, (key<<1)+d)
	}

}
func (self *snap) check(can config.Candles) bool {

	diff := can.GetMidAverage() - self.ca.GetMidAverage()
	if math.Abs(diff) > self.diff {
		self.forecast = (diff>0 == self.direction)
		return true
	}
	return false
}
type snaps struct{
	snaps []*snap
	//Merge [2]float64
}
func (self *snaps) Append(sn *snap) {
	self.snaps = append(self.snaps,sn)
}
func (self *snaps) ratio() bool {

	for _, sn := range self.snaps {
		if !sn.forecast {
			return false
		}
	}
	return true

}
func (self *snaps) Len() int {
	return len(self.snaps)
}

type JointCache struct {

	EndJoint  *Joint
	BeginJoint *Joint
	IsSplit bool

	TmpSnap *snaps
	SnapMap map[uint64]*snaps

	Cov [2]Cov
	Merge [2]float64
	parSignal *Signal

}

func NewJointCache(par *Signal) *JointCache {

	jc := JointCache{EndJoint:new(Joint),
			TmpSnap:new(snaps),
			SnapMap:make(map[uint64]*snaps),
			parSignal:par}
	jc.BeginJoint = jc.EndJoint
	return &jc

}

func (self *JointCache) AddSnapMap(sn *snap){

	var sns *snaps
	var _sn,beginSn *snap
	for _,k := range sn._key {
		sns = self.SnapMap[k]
		if sns == nil {
			self.SnapMap[k] = &snaps{snaps:[]*snap{sn}}
		}else {
			sns.Append(sn)

			_sn = sns.snaps[0]
			if beginSn == nil {
				beginSn = _sn
			}else if beginSn.ca.GetTime()> _sn.ca.GetTime() {
				beginSn = _sn
			}
		}
	}

	if beginSn == nil {
		return
	}
	if beginSn.ca.GetTime() > self.BeginJoint.Cans[0].GetTime() {
		return
	}
	var I int
	for _,k := range beginSn._key {
		sns = self.SnapMap[k]
		if sns==nil {
			continue
		}
		for _i,_sn := range sns.snaps {
			if _sn == beginSn {
				I =_i+1
				if I == len(sns.snaps) {
					delete(self.SnapMap,k)
				}else{
					sns.snaps = sns.snaps[I:]
				}
				continue
			}
		}
	}

}
func (self *JointCache) CheckSnap(sn *snap){

	if !sn.start {
		return
	}
	sn.Post = true
	self.parSignal.ReadJoint(func(jc *JointCache,i int){
		if !sn.Post {
			return
		}
		for _,_sn := range jc.TmpSnap.snaps {
			if sn.direction != _sn.direction  {
				sn.Post = false
				return
			}
		}
	})

}

func (self *JointCache) CheckTmpSnap(can config.Candles){

	NewSnap:= make([]*snap,self.TmpSnap.Len())
	i := 0
	for _,sn := range self.TmpSnap.snaps {
		if !sn.check(can) {
			NewSnap[i] = sn
			i++
		}else{
			if sn.start {
				//if sn.Post {
					if sn.forecast {
						self.parSignal.MsgBox[0]++
					}else{
						self.parSignal.MsgBox[1]++
					}
				//}
			}
		}
	}
	self.TmpSnap.snaps = NewSnap[:i]

}
func (self *JointCache) AddTmpSnap (_sn *snap) {

	//self.CheckSnap(_sn)
	if _sn.start {
		self.TmpSnap.Append(_sn)
	}

}
func (self *JointCache) Add(ca config.Cache){
	can := ca.GetlastCan()
	self.CheckTmpSnap(can)
	self.EndJoint, self.IsSplit = self.EndJoint.Append(can,ca.GetScale())
	if !self.IsSplit {
		return
	}
	self.ClearJoint()
}
func (self *JointCache) ClearJoint(){
	if self.EndJoint.Last == nil {
		return
	}
	if self.EndJoint.Last.Last == nil {
		return
	}
	fd := self.EndJoint.Diff > 0
	if fd != (self.EndJoint.Last.Last.Diff>0) {
		panic(0)
	}
	var id,id1 int
	if fd {
		id = 1
		id1= 0
	}else{
		id = 0
		id1= 1
	}
	Cov := self.Cov[id].DummyAdd(self.EndJoint.Last.Last)
	if Cov == nil {
		return
	}
	self.Cov[id] = *Cov
	//self.Cov[id].AppendJos(self.EndJoint.Last.Last)
	//fmt.Println(self.BeginJoint.Scale,Cov.GetVal(),Cov.GetCount() , config.Conf.JointMax)
	var endjo,_jo *Joint
	for (Cov.GetCount() > config.Conf.JointMax) {
		_jo = Cov.Del()
		//fmt.Println(Cov.GetVal(),self.Cov[id].GetVal())
		if math.Abs(Cov.GetVal()) <= math.Abs(self.Cov[id].GetVal()) {
			self.Cov[id] = *Cov
			endjo = _jo
		}
	}
	//fmt.Println(self.BeginJoint.Scale,self.EndJoint.Num - self.BeginJoint.Num,Cov.GetVal())
	if endjo == nil {
		return
	}
	if endjo.Next == nil {
		return
	}
	if len(endjo.Next.Cans) == 0 {
		return
	}
	self.BeginJoint = endjo.Next

	self.BeginJoint.Cut()
	Jos:=self.Cov[id1].GetJoints()
	var index int = -1
	for i,jo := range Jos {
		if self.BeginJoint == jo {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	self.Cov[id1].Dels(index)


}
