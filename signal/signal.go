package signal
import(
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"fmt"
	"sync"
	"math"
	//"time"
	//"strings"
)
func init(){
	replay.SignalGroup = append(replay.SignalGroup, NewAna())
}
type Msg struct {

	//InsName string

	joint *replay.Joint
	can *request.Candles
	//takeProfit *request.Candles
	//stopLosses *request.Candles
	tp float64
	sl float64
	scale int64
	isSend bool

	last *Msg

}

func (self *Msg) test(pip float64) ( out bool) {

	joint,index:= self.joint.Find(self.can)
	if index <0 {
		panic("index")
	}
	val:=self.can.GetMidAverage()
	if joint == self.joint ||
	(((val - joint.Cans[0].GetMidAverage())>0) ==(self.tp>0)) {
		if joint.Next != nil &&
		joint.Next.Next != nil &&
		joint.Next.Next.Next != nil {
			Ncans:=joint.Next.Next.Next.Cans
			diff := Ncans[len(Ncans)-1].GetMidAverage() - val
			if (diff>0)==(self.tp>0) {
				replay.SignalBox[0]++
				replay.SignalBox[2]+=(math.Abs(diff) - pip)
				return true
			}else if (diff>0)==(self.sl>0) {
				replay.SignalBox[1]++
				replay.SignalBox[2]-=(math.Abs(diff) + pip)
				return true
			}
		}
	}else{
		if joint.Next != nil &&
		joint.Next.Next != nil &&
		joint.Next.Next.Next != nil &&
		joint.Next.Next.Next.Next != nil {
			Ncans:=joint.Next.Next.Next.Next.Cans
			diff := Ncans[len(Ncans)-1].GetMidAverage() - val
			if (diff>0)==(self.tp>0) {
				replay.SignalBox[0]++
				replay.SignalBox[2]+=(math.Abs(diff) - pip)
				return true
			}else if (diff>0)==(self.sl>0) {
				replay.SignalBox[1]++
				replay.SignalBox[2]-=(math.Abs(diff)+ pip)
				return true
			}

		}
	}
	//tp:=self.takeProfit.GetMidAverage() - val
	//sl:=self.stopLosses.GetMidAverage() - val
	var diff float64
	joint.Read(index+1,func(can *request.Candles) bool{
		if can.Time < self.can.Time {
			panic("can time")
		}
		diff = can.GetMidAverage() - val
		if ((diff>0)==(self.tp>0)) && (math.Abs(diff)>math.Abs(self.tp)) {
			replay.SignalBox[0]++
			replay.SignalBox[2]+=(math.Abs(diff) - pip)
			//fmt.Println(self.tp,self.sl,diff,"tp")
			out = true
			return true
		}else if ((diff>0)==(self.sl>0)) && (math.Abs(diff)>math.Abs(self.sl)) {
			out = true
			//fmt.Println(self.tp,self.sl,diff,"sl")
			replay.SignalBox[1]++
			replay.SignalBox[2]-=(math.Abs(diff)+pip)
			return true
		}
		return false

	})

	return out
}
func (self *Msg) check (s *Msg) {
	if self.joint.Next == nil {
		s.last = self
		s = self
	}
	last := self.last
	if last != nil {
		self.last = nil
		last.check(s)
	}

}
func (self *Msg) checkSignal( num *int) {

	if (self.last == nil){
		return
	}

	if (self.last.joint.MaxDiff !=0) ||
	((self.joint.Diff>0) != (self.last.joint.Diff>0)) {
		*num = 0
		return
	}
	*num++
	self.last.checkSignal(num)

}
type Signal struct {
	Msg *Msg
	Send chan *Msg
	InsCache *replay.InstrumentCache
}
func (self *Signal) readSend( f func(*Msg,float64) bool){
	le := len(self.Send)
	for i:=0;i<le;i++{
		mg:=<-self.Send
		//if !f(mg,self.InsCache.Ins.PipDiff()){
		if !f(mg,0){
			self.Send<-mg
		}
	}
}

type Ana struct{
	//msg
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
	if sig == nil  {
		self.Lock()
		self.msgMap[key] = &Signal{
			Msg:msg,
			InsCache:Ins,
			Send:make(chan *Msg,100)}
		self.Unlock()
		return
	}
	sig.Msg.check(msg)
	sig.Msg = msg
	if msg.joint.Last != nil &&
	msg.joint.Last.Last != nil &&
	(math.Abs(msg.joint.Last.Last.Diff)/math.Abs(msg.joint.Diff)) > 3 {
		var num int = 0
		msg.checkSignal(&num)
		if num >0 {
			select{
			case sig.Send <-msg:
				msg.isSend = true
			default:
			}
		}
	}
	//self.msgMap[key] = signalmsg

}
func (self *Ana) Check(InsCache *replay.InstrumentCache){
	f := func(Cache *replay.Cache,CacheId int) {
		dif :=(Cache.EndJoint.Diff >0)
		Cache1 := InsCache.CacheList[CacheId+1]
		if (Cache1.EndJoint.MaxDiff != 0){
			return
		}
		if (dif != (Cache1.EndJoint.Diff>0)) {
			return
		}
		Cache2 := InsCache.CacheList[CacheId+2]
		if (Cache2.EndJoint.MaxDiff != 0){
			return
		}
		if (dif != (Cache2.EndJoint.Diff>0)) {
			return
		}

		sl := Cache1.EndJoint.Cans[0].GetMidAverage() - Cache.LastCan.GetMidAverage()
		if math.Abs(sl) < Cache1.EndJoint.GetLongAve(){
			return
		}
		self.add(&Msg{
			tp:-sl,
			sl:sl,
			joint:Cache.EndJoint,
			isSend:false,
			scale: Cache.Scale,
			can:Cache.LastCan},InsCache)
		//fmt.Println(sl,-sl)
	}

	sig := self.msgMap[InsCache.Ins.Name]
	s := func(Cache *replay.Cache) {
		if sig == nil {
			return
		}
		sig.readSend(func(mg *Msg,pip float64) bool{
			//fmt.Println(pip)
			if mg.scale == Cache.Scale &&
			mg.test(pip) {
				return true
			}else{
				return false
			}
		})
	}

	for i := len(InsCache.CacheList)-3;i>=0;i--{
		tmpca:=InsCache.CacheList[i]
		if tmpca.IsSplit {
			s(tmpca)
			f(tmpca,i)
		}
	}

}
func (self *Ana) Show(){
	fmt.Println("Hello world")
}
