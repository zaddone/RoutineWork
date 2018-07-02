package signal

import(
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"fmt"
	"sync"
	"math"
	"log"
	"strconv"
	//"time"
	"strings"
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
	mr *request.OrderResponse

}
func (self *Msg) GetTime() int64 {
	return self.can.Time + self.scale + self.scale
}
func (self *Msg) Send(InsName string){
	return
	mr,err := request.HandleOrder(InsName,config.Conf.Units,0,self.tp,self.sl)
	if err != nil {
		log.Println(err)
	}else{
		self.mr = &mr
		log.Println(self.mr)
	}
}
func (self *Msg) Close(diff float64) {

	//diff := Ins.CacheList[0].LastCan.GetMidAverage() - self.can.GetMidAverage()
	//diff := LastCanVal - self.can.GetMidAverage()
	if (diff>0) == (self.tp>0) {
		replay.SignalBox[0]++
		replay.SignalBox[2]+=math.Abs(diff)
	}else{
		replay.SignalBox[1]++
		replay.SignalBox[2]-=math.Abs(diff)
	}

}

func (self *Msg) test(Cache *replay.Cache) ( out bool) {

	val:=self.can.GetMidAverage()
	var diff float64
	out = false

	joint,index:= self.joint.Find(self.can)
	if index <0 {
		panic("index")
	}
	//lastCan := self.can
	//var tdiff int64
	joint.Read(index+1,func(can *request.Candles) bool{

		//tdiff = (can.Time - lastCan.Time)
		//if tdiff > self.scale*10 {
		//	fmt.Println(self.scale,tdiff,"---------------------------")
		//}
		//lastCan = can

		if can.Time < self.can.Time {
			panic("can time")
		}
		diff = can.GetMidAverage() - val
		//if ((diff>0)==(self.tp>0)) && (math.Abs(diff)>math.Abs(self.tp)) {
		//	replay.SignalBox[0]++
		//	replay.SignalBox[2]+=math.Abs(diff)
		//	out = true
		//	return true
		//}else if ((diff>0)==(self.sl>0)) && (math.Abs(diff)>math.Abs(self.sl)) {
		if ((diff>0)==(self.sl>0)) && (math.Abs(diff)>math.Abs(self.sl)) {
			out = true
			replay.SignalBox[1]++
			replay.SignalBox[2]-= math.Abs(diff)
			return true
		}
		return false

	})
	if !out {
		tp:=(self.tp>0)
		if  tp != (Cache.EndJoint.Diff>0){
			diff = (Cache.LastCan.GetMidAverage() - val)
			if tp == (diff>0) {
				replay.SignalBox[0]++
				replay.SignalBox[2]+= math.Abs(diff)
				out = true
				//return true
				if self.mr != nil {
					v,err := request.ClosePosition(Cache.GetInsName(),fmt.Sprintf("%d",config.Conf.Units))
					if err != nil {
						log.Println(err)
					}else{
						if v>0 {
							replay.SignalBox[4]++
						}else{
							replay.SignalBox[5]++
						}
						replay.SignalBox[3]+=v
					}
				}
			}
		}
	}

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
func (self *Msg) checkSignal( num *int,msg *Msg) {

	if self.scale >= msg.scale {
		//if (self.joint.MaxDiff !=0) ||
		if ((msg.joint.Diff>0) != (self.joint.Diff>0)) {
			*num = 0
			return
		}
		*num++
	}
	if self.last != nil {
		self.last.checkSignal(num,msg)
	}

}
type Signal struct {
	Msg *Msg
	Send []*Msg
	InsCache *replay.InstrumentCache
}
func (self *Signal) CheckMsg(msg *Msg) bool {

	self.Msg.check(msg)
	self.Msg = msg
	if math.Abs(msg.sl)<self.InsCache.Ins.MinimumTrailingStopDistance*1.5 {
		return false
	}

	if (len(self.Send) >0) {
		if (self.Send[0].tp>0) != (msg.tp>0) {
			return false
		}
	}else{
		if msg.last == nil {
			return false
		}
		var num int = 0
		msg.last.checkSignal(&num,msg)
		if num == 0 {
			return false
		}
	}

	msg.isSend = true
	//sig.CheckSend(msg)
	self.Send = append(self.Send,msg)

	//msg.GetTime()
	func(){
		if len(self.InsCache.Price.Time) == 0 {
			return
		}
		ti := strings.Split(string(self.InsCache.Price.Time),".")
		sec,err := strconv.Atoi(ti[0])
		if err != nil {
			log.Println(err)
			return
		}
		if msg.GetTime() < int64(sec) {
			return
		}
		for _,b := range self.InsCache.Price.Bids {
			if !msg.can.CheckSection(b.Price.GetPrice()) {
				return
			}
		}
		for _,b := range self.InsCache.Price.Asks {
			if !msg.can.CheckSection(b.Price.GetPrice()) {
				return
			}
		}
		units :=config.Conf.Units
		if msg.tp<0 {
			units = -units
		}
		mr,err := request.HandleOrder(self.InsCache.Ins.Name,units,0,msg.tp*2,msg.sl)
		if err != nil {
			log.Println(err)
			return
		}
		msg.mr = &mr

		log.Println(mr)
		//if msg.tp>0 {
		//	self.InsCache.Price.CloseoutBid
		//}else{
		//	self.InsCache.Price.CloseoutAsk
		//}
	}()

	return true

}

func (self *Signal) readSend( f func(*Msg) bool){
	NewSend := make([]*Msg,len(self.Send))
	j:=0
	for _,mg := range self.Send {
		if !f(mg){
			NewSend[j] = mg
			j++
		}
	}
	self.Send = NewSend[:j]
}
func (self *Signal) CheckSend(mg *Msg) {

	le := len(self.Send)
	if le == 0 {
		return
	}
	le--
	lastMsg :=self.Send[le]
	tp :=lastMsg.tp>0
	if tp != (mg.tp>0) {
		diff := (self.InsCache.GetBaseCan().GetMidAverage() - lastMsg.can.GetMidAverage())
		if tp != (diff>0) {
			lastMsg.Close(diff)
			self.Send = self.Send[:le]
			self.CheckSend(mg)
		}
		mg.isSend = false
	}

	//return
	//if len(self.Send) >0 && (self.Send[0].tp>0) != (mg.tp>0) {
	//	self.Send[0].Close(self.InsCache)
	//	self.Send = self.Send[1:]
	//	self.CheckSend(mg)
	//}

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
			Send:make([]*Msg,0,100)}
		self.Unlock()
		return
	}
	if sig.CheckMsg(msg) {
		msg.Send(sig.InsCache.Ins.Name)
	}

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
		sig.readSend(func(mg *Msg) bool{
			if (mg.scale == Cache.Scale) &&
			mg.test(Cache) {
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
