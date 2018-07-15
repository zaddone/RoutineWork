package replay

import (
	"log"
//	"flag"
	"strings"
	"github.com/zaddone/RoutineWork/config"
	//"github.com/zaddone/RoutineWork/request"
)

var (
	InsCaches []*InstrumentCache
	//InsName   = flag.String("n", "EUR_USD", "INS NAME")
	SignalGroup []Signal = nil
	SignalBox [6]float64
)

type Signal interface{
	Check(sd *InstrumentCache)
	Show()
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
