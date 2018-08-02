package replay

import (
	"log"
	"strings"
	"github.com/zaddone/RoutineWork/config"
	//"github.com/zaddone/RoutineWork/request"
)

var (
	InsCaches []*InstrumentCache
	SignalGroup []Signal = nil
	//SignalBox [6]float64
	//JointLibs []JointLib = nil
)

type Signal interface{
	//New()
	Update(config.Cache)
	Check(config.Instrument)
	Show()

}

func runSignal( f func(Signal)){

	if SignalGroup == nil  {
		return
	}
	for _,_s := range SignalGroup {
		f(_s)
		//si.Check(self)
	}

}
//type JointLib interface {
//	New()
//	Update(*request.Candles) bool
//}

func init() {
	log.Println("start replay")
	//flag.Parse()
	nas := strings.Split(config.Conf.InsName, "|")
	InsCaches = make([]*InstrumentCache, len(nas))
	var InsC *InstrumentCache
	for i, na := range nas {
		//InsC = new(InstrumentCache)
		InsC = NewInstrumentCache(na)
		if InsC == nil {
			log.Println(na,"Not fount")
			continue
		}
		InsCaches[i] = InsC
		//InsC.Init(na)
		InsC.Run()
	}
}

func Start(Ins string) bool {
	for _, insc := range InsCaches {
		if insc.Ins.Name == Ins {
			return false
		}
	}
	InsC := NewInstrumentCache(Ins)
	if InsC == nil {
		log.Println(Ins,"Not fount")
		return false
	}
	//InsC.Init(Ins)
	InsCaches = append(InsCaches, InsC)
	InsC.Run()
	log.Println("start", Ins)
	return true

}
