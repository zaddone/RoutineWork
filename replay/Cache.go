package replay

import (
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	id int
	Scale       int64
	Name        string
	FutureChan  chan *CacheFile
	EndtimeChan chan int64
	//DiffLong  float64
	LastCan   config.Candles

//	EndJoint  *Joint
//	BeginJoint *Joint
//	IsSplit bool

	//LastCache config.Cache
	par       config.Instrument
}
func (self *Cache) GetId()int{
	return self.id
}
func (self *Cache) SetId(i int){
	self.id = i
}
//self (self *Cache) GetBeginTime() int64 {
//	ca := self.par.GetEndCache()
//	if ca == self {
//		return 0
//	}
//	if ca.
//}

//func (self *Cache) SetLastCache(c config.Cache){
//	self.LastCache = c
//}
func (self *Cache) GetName() string {
	return self.Name
}
func (self *Cache) GetlastCan() config.Candles {
	return self.LastCan
}

func (self *Cache) GetInsCache() config.Instrument {
	return self.par
}

func (self *Cache) GetMinCache() config.Cache {
	return self.GetInsCache().GetBaseCache()
}

func (self *Cache) GetMinCan() config.Candles {
	return self.GetMinCache().GetlastCan()
}

func (self *Cache) GetInsName() string {
	return self.GetInsCache().GetInsName()
}

func NewCache(name string, scale int64, p *InstrumentCache) (ca *Cache) {

	ca = &Cache{
		Name:        name,
		Scale:       scale,
		FutureChan:  make(chan *CacheFile, 10),
		//EndJoint:    new(Joint),
		EndtimeChan: make(chan int64, 1),
		par:         p}
	//ca.BeginJoint = ca.EndJoint
	//ca.JointLib.New()
	return ca

}

func (self *Cache) Init(name string, scale int64, p *InstrumentCache) {
	self.Name = name
	self.Scale= scale
	self.FutureChan = make(chan *CacheFile, 10)
	self.EndtimeChan= make(chan int64, 1)
	self.par = p
}

func (self *Cache) GetScale() int64 {
	return self.Scale
}
func (self *Cache) Load(name, path string) {

	f, err := os.Stat(path)
	var cf *CacheFile
	if err == nil && f.IsDir() {
		filepath.Walk(path, func(pa string, fi os.FileInfo, er error) error {
			if fi.IsDir() {
				return er
			}
			//cf = new(CacheFile)
			//er = cf.Init(name, pa, fi, int(86400/self.Scale+1)*2)
			cf,er = NewCacheFile(name, pa, fi, int(86400/self.Scale+1)*2)
			if er == nil {
				self.FutureChan <- cf
			} else {
				fmt.Println(er)
			}
			return er
		})

	}
	var begin int64
	if cf == nil {
		beginT, err := time.Parse("2006-01-02T15:04:05", config.Conf.BEGINTIME)
		if err != nil {
			panic(err)
		}
		begin = beginT.UTC().Unix()
	} else {
		begin = cf.EndCan.GetTime() + self.Scale
	}
	cf = &CacheFile{Can:make(chan config.Candles, 1000)}
	//cf.Can = make(chan *request.Candles, 1000)
	self.FutureChan <- cf
	request.Down(name, begin, 0, self.Scale, self.Name, func(can *request.Candles) {
		cf.Can <- can
	})
	close(cf.Can)

}

func (self *Cache) CheckUpdate(can config.Candles) bool {

	if can.GetMidLong() == 0 {
		return false
	}
	if self.LastCan != nil {
		if self.LastCan.GetTime() >= can.GetTime() {
			return false
		}
	}
	self.LastCan = can
	//fmt.Println(can)
	return true

}
func (self *Cache) SetEndTimeChan(ti int64) {
	//self.IsSplit = false
	self.EndtimeChan <- ti

}
func (self *Cache) Sensor(cas []config.Cache) {
	calen := len(cas)
	//self.IsUpdate = false
	//var date string = ""
	self.Read(func(can config.Candles) {
		if !self.UpdateJoint(can) {
			return
		}
		endTime := can.GetTime() + self.Scale
		self.par.GetWait().Add(calen)
		for _, ca := range cas {
			go ca.SetEndTimeChan(endTime)
		}
		self.par.GetWait().Wait()
		//fmt.Printf("%s \r",time.Unix(endTime,0))

		//day:= time.Unix(endTime, 0)
		////d := day.Format("2006-01-02")
		//d := day.Format("2006-01")
		//if d != date {
		//	SignalBox[3] = SignalBox[1]/SignalBox[0]
		//	fmt.Printf("%s %.5f---------------------\r\n",date,SignalBox)
		//	date = d
		//	SignalBox=[6]float64{0,0,0,0,0,0}
		//}

		runSignal(func(s Signal){
			s.Check(self.GetInsCache())
		})

	})
}

func (self *Cache) SyncRun(hand func(can config.Candles) bool) {
	self.Read(func(can config.Candles){
		//self.IsUpdate = false
		for{
			endTime := <-self.EndtimeChan
			if can.GetTime()+self.Scale <= endTime {
				hand(can)
				//fmt.Println(self.Name)
				//self.IsUpdate = true
				//self.par.w.Done()
				self.EndtimeChan <- endTime
				return
			}
			self.par.GetWait().Done()
		}
	})
}

func (self *Cache) UpdateJoint(can config.Candles) ( bool) {

	if !self.CheckUpdate(can) {
		return false
	}
	self.GetInsCache().Monitor(self,can)
	//self.JointLib.Update(can)
	runSignal(func(s Signal){
		s.Update(self)
	})
	return true

//	self.EndJoint, self.IsSplit = self.EndJoint.Append(can)
//	//fmt.Println(len(self.par.SplitCache))
//	if self.IsSplit {
//		if self.LastCache == nil {
//			j :=0
//			self.EndJoint.ReadLast(func(jo *Joint) bool {
//				j++
//				if j < 4 {
//					return false
//				}
//				jo.Cut()
//				self.BeginJoint = jo
//				//fmt.Println("end cache",self.Name,time.Unix(jo.Cans[0].Time,0))
//				return true
//			})
//		} else {
//			//self.par.SplitCache <- self
//			if len(self.LastCache.BeginJoint.Cans) > 0 {
//				endTime := self.LastCache.BeginJoint.Cans[0].Time
//				self.BeginJoint.ReadNext(func(jo *Joint) bool {
//					if len(jo.Cans) > 0 && jo.Cans[0].Time < endTime {
//						return false
//					}
//					if self.BeginJoint != jo {
//						self.BeginJoint = jo
//						jo.Cut()
//						//fmt.Println(self.Name,self.BeginJoint.Cans[0].Time,endTime)
//					}
//					return true
//				})
//			}
//		}
//	}


}

func (self *Cache) Read(Handle func(can config.Candles)) {

	for {
		cf := <-self.FutureChan
		if cf == nil {
			break
		}
		for {
			can := <-cf.Can
			if can == nil {
				break
			}
			Handle(can)
		}
	}

}

