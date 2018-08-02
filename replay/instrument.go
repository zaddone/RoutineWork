package replay

import (
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"log"
	"path/filepath"
	"sync"
	"encoding/json"
	"fmt"
	"strings"
	"strconv"
)

type InstrumentCache struct {

	cacheLen	int
	cacheList       []config.Cache
	ServerChanMap   *ServerChanMap
	w		sync.WaitGroup
	Ins		*request.Instrument
	Price		request.Price

}

func NewInstrumentCache(Instr string) (inc *InstrumentCache) {

	Ins := request.ActiveAccount.Instruments[Instr]
	if Ins == nil {
		return nil
	}
	inc = &InstrumentCache{
		ServerChanMap:NewServerChanMap(),
		Ins:Ins,
		cacheList:make([]config.Cache,len(config.Conf.Granularity))}
	i := 0
	for k, v := range config.Conf.Granularity {
		ca := NewCache(k, int64(v), inc)
		go ca.Load(Ins.Name, filepath.Join(Ins.Name, ca.Name))
		//inc.CacheList = append(inc.CacheList, ca)
		inc.cacheList[i] = ca
		inc.Sort(i)
		i++
	}
	inc.GetCacheList(func(i int,c config.Cache){
		c.SetId(i)
	})
	//le := len(inc.cacheList) - 1
	//self.endCache := inc.cacheList[le]
	if config.Conf.Price {
		go inc.syncGetPrice()
	}
	return

}
func (self *InstrumentCache) GetCacheLen() int {
	return len(self.cacheList)
}
func (self *InstrumentCache) GetCache(id int) config.Cache {
	return self.cacheList[id]
}
func (self *InstrumentCache) GetCacheList (f func (int,config.Cache)) {

	for i,c := range self.cacheList {
		f(i,c)
	}
}
func (self *InstrumentCache)GetStandardPrice(p float64) string {
	return self.Ins.StandardPrice(p)
}
func (self *InstrumentCache) GetPriceTime() int64 {
	if len(self.Price.Time) == 0 {
		return 0
	}
	ti := strings.Split(string(self.Price.Time),".")
	if len(ti) == 0 {
		return 0
	}
	sec,err := strconv.Atoi(ti[0])
	if err != nil {
		log.Println(err)
		return 0
	}
	return int64(sec)

}
func (self *InstrumentCache) GetEndCache() config.Cache {

	return self.cacheList[len(self.cacheList)-1]

}
func (self *InstrumentCache) GetWait() *sync.WaitGroup{
	return &self.w
}

func (self *InstrumentCache) Monitor(ca config.Cache,can config.Candles) {

	self.ServerChanMap.Send(&TimeCache{
		Scale: ca.GetScale(),
		Time:  can.GetTime(),
		Name:  ca.GetName()})

}
func (self *InstrumentCache) GetInsStopDis() float64 {
	return self.Ins.MinimumTrailingStopDistance
}
func (self *InstrumentCache) GetInsName() string{
	return self.Ins.Name
}

func (self *InstrumentCache) syncGetPrice(){

	var err error
	for{
		err = request.GetPricingStream([]string{self.GetInsName()},func(da []byte){
			if len(da) == 0 {
				return
			}
			err = json.Unmarshal(da,&self.Price)
			if err != nil {
				log.Println(string(da),err)
			}
			fmt.Printf("%s\r",self.Price.Time)
		})
	}

}
func (self *InstrumentCache) Sort(i int){
	if i == 0 {
		return
	}
	I := i - 1
	if self.cacheList[I].GetScale() > self.cacheList[i].GetScale() {
		self.cacheList[I], self.cacheList[i] = self.cacheList[i], self.cacheList[I]
		self.Sort(I)
	}
}

func (self *InstrumentCache) Run(){

	cas:=self.cacheList[1:]
	for _, _ca := range cas {
		go _ca.SyncRun(_ca.UpdateJoint)
	}
	go self.cacheList[0].Sensor(cas)

}
func (self *InstrumentCache) GetBaseCache() config.Cache {
	return self.cacheList[0]
}
func (self *InstrumentCache) GetHight(ca config.Cache) config.Cache {
	le := len(self.cacheList)-1
	for i:=0;i<le;i++ {
		if self.cacheList[i] == ca {
			return self.cacheList[i+1]
		}
	}
	return nil
}

//func (self *InstrumentCache) GetBaseCan() *request.Candles {
//	return self.GetBaseCache().LastCan
//}

//func (self *InstrumentCache) Selective(f bool,begin int) (end int) {
//
//	end = begin+1
//	if end == len(self.CacheList) {
//		return begin
//	}
//
//	Cache := self.CacheList[end]
//	if (Cache.EndJoint.MaxDiff != 0) {
//		if (f == (Cache.EndJoint.Diff>0)) {
//			return begin
//		}
//	}else{
//		if (f != (Cache.EndJoint.Diff>0)) {
//			return begin
//		}
//	}
//	return self.Selective(f,end)
//
//}

