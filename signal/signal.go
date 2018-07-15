package signal

import (
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"fmt"
	"math"
	"log"
	"strconv"
	"time"
	"strings"
)

type Signal struct {
	Msg *Msg
	InsCache *replay.InstrumentCache
}

func (self *Signal) Add(msg *Msg) {

	if self.Msg == nil {
		self.Msg = msg
		return
	}
	self.Msg = self.Msg.merge(msg)
	//	return
	//}
	//if !self.Msg.post{
	//	msg.last = self.Msg.last
	//	self.Msg = msg
	//	return
	//}
	//msg.last = self.Msg
	//self.Msg = msg
	return

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
	if ( math.Abs(msg.sl) < self.InsCache.Ins.MinimumTrailingStopDistance*1.5 ) {
		return
	}
	//fmt.Println(time.Unix(msg.can.Time,0),msg.endCache.Scale,msg.sl)
	msg.post = true
	//msg.Show(0)
	if msg.last != nil {
		msg.last = msg.last.readAll( func(m *Msg) bool {
			m.Close(self.InsCache.GetBaseCan().GetMidAverage())
			return true
		})
		//fmt.Println("last",msg.last)
		//msg.last.Close(self.InsCache.GetBaseCan().GetMidAverage())
		//msg.last = nil
	}

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
		fmt.Println(time.Unix(msg.GetTime(),0).Format("2006-01-02T15:04:05"),time.Unix(int64(sec),0).Format("2006-01-02T15:04:05"))
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

	sl:=self.InsCache.Ins.StandardPrice(beginpr + msg.sl)
	tp:=self.InsCache.Ins.StandardPrice(beginpr + msg.tp)
	fmt.Println(sl,tp)

	mr,err := request.HandleOrder(
		self.InsCache.Ins.Name,
		units,"",tp,sl)
	if err != nil {
		log.Println(err)
		return
	}
	msg.mr = &mr

	log.Println(mr)

}

