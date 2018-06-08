package replay

import (
	//"../request"
	"context"
	//"flag"
	//"log"
	"path/filepath"
	"sync"
	//"strings"
)


type TimeCache struct {
	Time  int64
	Name  string
	Scale int64
}

type ServerChan struct {
	TimeChan chan TimeCache
	ctx      context.Context
}

func (self *ServerChan) Out(f func(tc *TimeCache) error) (err error) {
	for {
		t := <-self.TimeChan
		err = f(&t)
		if err != nil {
			return err
		}

	}
	return err
}

func (self *ServerChan) In(f TimeCache) {
	ctx, _ := context.WithCancel(self.ctx)

	go func(_f TimeCache, _ctx context.Context) {
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
	self.TimeChan = make(chan TimeCache, 5000)
}

type ServerChanMap struct{
	ServerChans     map[int]*ServerChan
	sync.RWMutex
}
func (self *ServerChanMap) Init(){
	self.ServerChans = make(map[int]*ServerChan)
}
func (self *ServerChanMap) Del(k int){
	self.Lock()
	delete(self.ServerChans,k)
	self.Unlock()
}
func (self *ServerChanMap) Send(tc TimeCache){
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
	Name           string
	ServerChanMap     ServerChanMap
}

func NewInstrumentCache(Instr string) *InstrumentCache {
	var inc InstrumentCache
	inc.Init(Instr)
	return &inc
}

func (self *InstrumentCache) Init(Instr string) {
	self.ServerChanMap.Init()
	//self.ServerChan = make(map[int]*ServerChan)
	self.Name = Instr
	self.GranularityMap = make(map[string]int64)
	self.GranularityMap["S5"] = 5
	self.GranularityMap["S10"] = 10
	self.GranularityMap["S15"] = 15
	self.GranularityMap["S30"] = 30
	self.GranularityMap["M1"] = 60
	self.GranularityMap["M2"] = 60 * 2
	self.GranularityMap["M4"] = 60 * 4
	self.GranularityMap["M5"] = 60 * 5
	self.GranularityMap["M10"] = 600
	self.GranularityMap["M15"] = 60 * 15
	self.GranularityMap["M30"] = 60 * 30
	self.GranularityMap["H1"] = 3600
	self.GranularityMap["H2"] = 3600 * 2
	self.GranularityMap["H3"] = 3600 * 3
	self.GranularityMap["H4"] = 3600 * 4
	self.GranularityMap["H6"] = 3600 * 6
	self.GranularityMap["H8"] = 3600 * 8
	self.GranularityMap["H12"] = 3600 * 12
	self.GranularityMap["D"] = 3600 * 24
	//GranularityMap["W"] = 3600 * 24 * 7
	//GranularityMap["M"]= 3600*24*30
	i := 0
	for k, v := range self.GranularityMap {
		ca := NewCache(k, v, self)
		//ca.Id = i
		go ca.Load(Instr, filepath.Join(Instr, ca.Name))
		self.CacheList = append(self.CacheList, ca)
		self.sortCacheList(i)
		i++
	}
	le := len(self.CacheList) - 1
	lastCache := self.CacheList[le]
	for _, ca := range self.CacheList[:le] {
		ca.lastCache = lastCache
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
	for _, _ca := range self.CacheList[1:] {
		go _ca.SyncRun(_ca.UpdateJoint)
	}
	go self.CacheList[0].Sensor(self.CacheList[1:])

}
