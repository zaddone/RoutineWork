package signal
import(
	//"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"math"
	"log"
	"fmt"
)
type Msg struct {

	joint *Joint
	can config.Candles
	endJoint *Joint
	endCache config.Cache
	tp float64
	sl float64
	//scale int64
	last *Msg
	mr *request.OrderResponse
	post bool
	Num int
	par  *Signal

}
//func NewMsg(s *Signal) *Msg {
//}

func (self *Msg) readAll(f func(m *Msg) bool) *Msg {

	if self.last!= nil {
		self.last = self.last.readAll(f)
	}
	if f(self) {
		return self.last
	}else{
		return self
	}

}

func (self *Msg) CheckIntersection(m *Msg) bool {

	val  := self.can.GetMidAverage()
	mval := m.can.GetMidAverage()

	s_tp :=val + self.tp
	s_sl :=val + self.sl

	if math.Max(s_tp,s_sl) < mval ||
	math.Min(s_tp,s_sl) > mval {
		return false
	}

	m_tp :=mval + m.tp
	m_sl :=mval + m.sl

	if math.Max(m_tp,m_sl) < val ||
	math.Min(m_tp,m_sl) > val {
		return false
	}

	return true

}

func (self *Msg) merge(m *Msg) *Msg {

	if (m.tp>0) != (self.tp>0) {
		if self.post {
			m.last = self
			return m
		}
		if self.last != nil {
			return self.last.merge(m)
		}
		return m
	}

	if !self.CheckIntersection(m) {
		m.last = self.last
		return m
	}

	if !self.post {
		//if self.CheckIntersection(m){
		sl := self.can.GetMidAverage() + self.sl - m.can.GetMidAverage()
		if ((sl>0) == (m.sl>0)) &&
		(math.Abs(m.sl) < math.Abs(sl)) {
			m.sl = sl
			m.tp = -sl
			m.endJoint = self.endJoint
			m.endCache = self.endCache
			m.Num = self.Num +1
		}
		//}
		m.last = self.last
		return m

	}else{
		//if self.CheckIntersection(m){
		sl := m.can.GetMidAverage() + m.sl - self.can.GetMidAverage()
		if ((sl>0) == (self.sl>0)) &&
		(math.Abs(self.sl) < math.Abs(sl)) {
			self.sl = sl
			self.tp = -sl
			self.endJoint = m.endJoint
			self.endCache = m.endCache
			//self.Num++
		}
		//}
		return self
	}

}

func (self *Msg) Count() (n int) {
	self.readAll(func(m *Msg) bool {
		n++
		return false
	})
	return n
}

func (self *Msg) Show(i int){
	fmt.Println(i,self.endCache.GetScale(),self.joint.Diff)
	if self.last != nil {
		self.last.Show(i+1)
	}
}

func (self *Msg) SendClose() {
	if self.mr == nil {
		return
	}

	log.Println("close position")
	v,err := request.ClosePosition(self.endCache.GetName(),fmt.Sprintf("%d",config.Conf.Units))
	if err != nil {
		log.Println("close",err)
	}else{
		fmt.Println(v)
		//if v>0 {
		//	self.par.MsgBox[4]++
		//}else{
		//	self.par.MsgBox[5]++
		//}
		//self.par.MsgBox[3] += v
		//fmt.Println(self.par.MsgBox[3:])
	}
}

func (self *Msg) GetTime() int64 {
	return self.can.GetTime() + 5+ 5
}

func (self *Msg) Close(val float64) {

	//fmt.Println("Close start")
	if !self.post {
		return
	}
	//diff := val - self.can.GetMidAverage()
	//if (diff>0) == (self.tp>0) {
	//	replay.SignalBox[0]++
	//	replay.SignalBox[2] += math.Abs(diff)
	//}else{
	//	replay.SignalBox[1]++
	//	replay.SignalBox[2]-= math.Abs(diff)
	//}
	self.SendClose()

}

//func (self *Msg) testClose() bool {
//
//	tp:=(self.tp>0)
//	if tp == (self.endCache.EndJoint.Diff>0) {
//		return false
//	}
//	diff := (self.endCache.GetMinCan().GetMidAverage() - self.can.GetMidAverage())
//	if (diff>0) != tp {
//		return false
//	}
//	if self.post {
//		replay.SignalBox[0]++
//		replay.SignalBox[2]+= math.Abs(diff)
//	}
//	self.SendClose()
//	return true
//
//}

func (self *Msg) testOut(Cache config.Cache) bool {

	diff := Cache.GetlastCan().GetMidAverage() - self.can.GetMidAverage()
	if (math.Abs(diff)<math.Abs(self.sl)) {
		return false
	}
	if ( (diff>0) == (self.sl>0) ) {
		if self.post {
			//replay.SignalBox[1]++
			//replay.SignalBox[2]-= math.Abs(diff)
		}
	}else{
		if self.post {
			//replay.SignalBox[0]++
			//replay.SignalBox[2]+= math.Abs(diff)
		}
	}
	return true

}
