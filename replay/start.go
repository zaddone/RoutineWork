package replay

import (
	"log"
	"flag"
	"strings"
	"github.com/zaddone/RoutineWork/request"
)

var (
	InsCaches []*InstrumentCache
	InsName   = flag.String("n", "EUR_JPY", "INS NAME")
	//SignalSys []Signal = nil
)
type SignalData struct {

	Ty bool
	InsCache *InstrumentCache
	NowCache *Cache
	Can *request.Candles

}
type Signal interface{
	Check(da *SignalData)
	Show()
}

func init() {
	log.Println("start replay")
	flag.Parse()
	nas := strings.Split(*InsName, "|")
	InsCaches = make([]*InstrumentCache, len(nas))
	var InsC *InstrumentCache
	for i, na := range nas {
		//InsC = new(InstrumentCache)
		InsC = NewInstrumentCache(na)
		InsCaches[i] = InsC
		//InsC.Init(na)
		InsC.Run()
	}
}

func Start(Ins string) bool {
	for _, insc := range InsCaches {
		if insc.Name == Ins {
			return false
		}
	}
	InsC := NewInstrumentCache(Ins)
	//InsC.Init(Ins)
	InsCaches = append(InsCaches, InsC)
	InsC.Run()
	log.Println("start", Ins)
	return true

}
