package replay

import (
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"context"
	//"flag"
	"log"
	"path/filepath"
	"sync"
	"encoding/json"
	"fmt"
	//"time"
	//"strings"
	//"fmt"
)


type TimeCache struct {
	Time  int64
	Name  string
	Scale int64
}

type ServerChan struct {
	TimeChan chan *TimeCache
	ctx      context.Context
}

func (self *ServerChan) Out(f func(tc *TimeCache) error) (err error) {

	for {
		t := <-self.TimeChan
		err = f(t)
		if err != nil {
			return err
		}

	}
	return err

}

func (self *ServerChan) In(f *TimeCache) {
	ctx, _ := context.WithCancel(self.ctx)

	go func(_f *TimeCache, _ctx context.Context) {
		select {
		case <-_ctx.Done():
			//log.Print("done stop")
			return
		case self.TimeChan<-f:
			return
		}
	}(f, ctx)
}

func (self *ServerChan) Init(ctx context.Context) {
	self.ctx = ctx
	self.TimeChan = make(chan *TimeCache, 5000)
}

type ServerChanMap struct{
	ServerChans     map[int]*ServerChan
	sync.RWMutex
}
func NewServerChanMap() *ServerChanMap {
	var scm ServerChanMap
	scm.Init()
	return &scm
}
func (self *ServerChanMap) Init(){
	self.ServerChans = make(map[int]*ServerChan)
}
func (self *ServerChanMap) Del(k int){
	self.Lock()
	delete(self.ServerChans,k)
	self.Unlock()
}
func (self *ServerChanMap) Send(tc *TimeCache){

	self.Lock()
	for _,ser := range self.ServerChans {
		ser.In(tc)
	}
	self.Unlock()

}
func (self *ServerChanMap) Add(k int,Sc *ServerChan){

	self.Lock()
	self.ServerChans[k] = Sc
	self.Unlock()

}

type InstrumentCache struct {
	GranularityMap map[string]int64
	CacheList      []*Cache
	//Name           string
	ServerChanMap  ServerChanMap
	w		sync.WaitGroup
	Ins		*request.Instrument
	Price		request.Price
	//SplitCache     chan *Cache
	//signal Signal
}

func NewInstrumentCache(Instr string) *InstrumentCache {

	Ins := request.ActiveAccount.Instruments[Instr]
	if Ins == nil {
		return nil
	}
	var inc InstrumentCache
	inc.Init(Ins)
	//inc.signal = SignalSys
	return &inc
}

func (self *InstrumentCache) GetBaseCache() *Cache {
	return self.CacheList[0]
}

func (self *InstrumentCache) GetBaseCan() *request.Candles {
	return self.GetBaseCache().LastCan
}
func (self *InstrumentCache) Signal(){

	if SignalGroup == nil  {
		return
	}
	for _,si := range SignalGroup {
		si.Check(self)
	}

}
func (self *InstrumentCache) Monitor(ca *Cache,can *request.Candles) {

	self.ServerChanMap.Send(&TimeCache{
		Scale: ca.Scale,
		Time:  can.Time,
		Name:  ca.Name})

}

func (self *InstrumentCache) Selective(f bool,begin int) (end int) {

	end = begin+1
	if end == len(self.CacheList) {
		return begin
	}

	Cache := self.CacheList[end]
	if (Cache.EndJoint.MaxDiff != 0) {
		if (f == (Cache.EndJoint.Diff>0)) {
			return begin
		}
	}else{
		if (f != (Cache.EndJoint.Diff>0)) {
			return begin
		}
	}
	return self.Selective(f,end)

}

func (self *InstrumentCache) GetHightCache(ca *Cache) *Cache {
	for i,cache := range self.CacheList {
		if cache == ca {
			return self.CacheList[i+1]
		}
	}
	return nil
}

func (self *InstrumentCache) Init(Ins *request.Instrument) {

	self.ServerChanMap.Init()
	self.Ins=  Ins
	//self.ServerChan = make(map[int]*ServerChan)
	//self.Name = Instr
	self.GranularityMap = map[string]int64{
		"S5"  : 5,
		"S10" : 10,
		"S15" : 15,
		"S30" : 30,
		"M1"  : 60,
		"M2"  : 60*2,
		"M4"  : 60*4,
		"M5"  : 60*5,
		"M10" : 60*10,
		"M15" : 60*15,
		"M30" : 60*30,
		"H1"  : 3600,
		"H2"  : 3600 * 2,
		"H3"  : 3600 * 3,
		"H4"  : 3600 * 4,
		"H6"  : 3600 * 6,
		"H8"  : 3600 * 8,
		"H12" : 3600 * 12,
		"D"   : 3600 * 24}
	//GranularityMap["W"] = 3600 * 24 * 7
	//GranularityMap["M"]= 3600*24*30
	i := 0
	for k, v := range self.GranularityMap {
		ca := NewCache(k, v, self)
		//ca.Id = i
		go ca.Load(self.Ins.Name, filepath.Join(self.Ins.Name, ca.Name))
		self.CacheList = append(self.CacheList, ca)
		self.sortCacheList(i)
		i++
	}
	le := len(self.CacheList) - 1
	lastCache := self.CacheList[le]
	for _, ca := range self.CacheList[:le] {
		ca.LastCache = lastCache
	}

	if config.Conf.Price {
		go self.syncGetPrice()
	}

}
func (self *InstrumentCache) syncGetPrice(){

	var err error
	for{
		err = request.GetPricingStream([]string{self.Ins.Name},func(da []byte){
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
func (self *InstrumentCache) sortCacheList(i int) {

	if i == 0 {
		return
	}
	I := i - 1
	if self.CacheList[I].GetScale() > self.CacheList[i].GetScale() {
		self.CacheList[I], self.CacheList[i] = self.CacheList[i], self.CacheList[I]
		self.sortCacheList(I)
	}

}

func (self *InstrumentCache) Run() {

	//endId := len(self.CacheList) - 1
	cas:=self.CacheList[1:]
	for _, _ca := range cas {
		go func(ca *Cache){
			ca.SyncRun(ca.UpdateJoint)
		}(_ca)
	}
	go self.CacheList[0].Sensor(cas)

}
