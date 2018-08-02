package replay

import (
	"context"
	"sync"
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
	sync.Mutex
}
func NewServerChanMap() *ServerChanMap {
	return &ServerChanMap{ServerChans:map[int]*ServerChan{}}
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

