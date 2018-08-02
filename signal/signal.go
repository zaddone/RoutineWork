package signal

import (
	//"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"fmt"
	"math"
	"log"
	//"strconv"
	"time"
	//"strings"
)

type Signal struct {

	//Msg *Msg
	InsCache config.Instrument
	Joint *JointMap
	MsgBox [2]float64
	//MsgBox1 []float64
	//Split chan int
	CacheLen int
	snap  int64

	LastTime time.Time

}

func NewSignal(Ins config.Instrument) (s *Signal) {

	s = &Signal{
		Joint:NewJointMap(),
		InsCache:Ins}

	if Ins != nil {
		s.CacheLen=Ins.GetCacheLen()
	}
	return s

}

func (self *Signal) GetMergeRa(){

	var mer [2]float64
	self.ReadJoint(func(jc *JointCache,i int){
		if jc == nil {
			return
		}
		mer[0] += jc.Merge[0]
		mer[1] += jc.Merge[1]
	})
	fmt.Printf("%.0f %.3f\r\n",mer,mer[0]/mer[1])

}

func (self *Signal) GetSnap() (snap int64) {

	self.ReadJoint(func(jo *JointCache,i int){
		snap = snap<<1
		if jo== nil {
			return
		}
		if jo.IsSplit {
			snap++
		}
	})
	return snap

}

func (self *Signal) CheckSnap() (ids []int) {

	snap := self.GetSnap()
	if self.snap == snap {
		return
	}
	xor := self.snap^snap
	self.snap = snap
	//fmt.Printf("%b %b %b\r\n",snap,self.snap,xor)
	for i:=self.CacheLen-1; i>=0; i-- {
		if (xor|1 == xor) && (snap|1 == snap) {
			ids = append(ids,i)
		}
		xor = xor>>1
		snap= snap>>1
	}
	return

}

func (self *Signal) ReadJoint(f func(*JointCache,int)) {
	self.InsCache.GetCacheList(func(i int,c config.Cache) {
		f(self.Joint.Get(c.GetScale()),i)
	})
}
func (self *Signal) ShowMsg() string {
	return fmt.Sprintf("%.0f %.3f",self.MsgBox,self.MsgBox[0]/self.MsgBox[1])
}

func (self *Signal) UpdateJoint( ca config.Cache ) {

	Time := time.Unix(ca.GetlastCan().GetTime(),0)
	if self.LastTime.Unix() < Time.Unix() {
		if Time.Month() != self.LastTime.Month() {
			fmt.Println(self.LastTime,self.ShowMsg())
			self.MsgBox = [2]float64{0,0}
		}
		self.LastTime = Time
		fmt.Printf("%s %s\r",self.LastTime,self.ShowMsg())
	}
	k := ca.GetScale()
	j := self.Joint.Get(k)
	if j == nil {
		j = NewJointCache(self)
		self.Joint.Set(k,j)
	}
	j.Add(ca)

}

func (self *Signal) PostOrder(msg *Msg) {
	if msg.post {
		return
	}
	//if msg.Num ==0 {
	//	return
	//}

	//Min:=self.InsCache.Ins.MinimumTrailingStopDistance*1.5
	//Max:=self.InsCache.Ins.MinimumTrailingStopDistance*2
	//sl := math.Abs(msg.sl)
	if (math.Abs(msg.sl)<self.InsCache.GetInsStopDis()*1.5){
		return
	}
	//fmt.Println(time.Unix(msg.can.Time,0),msg.endCache.Scale,msg.sl)
	msg.post = true
	//msg.Show(0)
	if msg.last != nil {
		msg.last = msg.last.readAll( func(m *Msg) bool {
			m.Close(self.InsCache.GetBaseCache().GetlastCan().GetMidAverage())
			return true
		})
		//fmt.Println("last",msg.last)
		//msg.last.Close(self.InsCache.GetBaseCan().GetMidAverage())
		//msg.last = nil
	}

	sec := self.InsCache.GetPriceTime()
	if sec == 0 {
		return
	}

	if msg.GetTime() < sec {
		fmt.Println(time.Unix(msg.GetTime(),0).Format("2006-01-02T15:04:05"),time.Unix(sec,0).Format("2006-01-02T15:04:05"))
		return
	}

	beginpr := msg.can.GetMidAverage()
	tpf :=(msg.tp>0)
	//pr := self.InsCache.Price.GetPrice()
	//diff := pr - beginpr

	//if ((diff>0) != tpf ) && !msg.can.CheckSection(pr) {
	//	fmt.Println("diff",diff,msg.tp)
	//	return
	//}
	units := config.Conf.Units
	if !tpf {
		units = -units
	}

	sl:=self.InsCache.GetStandardPrice(beginpr + msg.sl)
	tp:=self.InsCache.GetStandardPrice(beginpr + msg.tp)
	fmt.Println(sl,tp)

	mr,err := request.HandleOrder(
		self.InsCache.GetInsName(),
		units,"",tp,sl)
	if err != nil {
		log.Println(err)
		return
	}
	msg.mr = &mr

	log.Println(mr)

}

