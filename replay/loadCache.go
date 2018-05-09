package replay

import (
	"../request"
	"bufio"
	"fmt"
	//	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//var (
//	GranularityMap map[string]int64
//	CacheList      []*Cache
//)

type CacheFile struct {
	Can    chan *request.Candles
	Fi     os.FileInfo
	Path   string
	EndCan *request.Candles
}

func (self *CacheFile) Init(name, path string, fi os.FileInfo, Max int) (err error) {

	self.Path = path
	self.Fi = fi
	self.Can = make(chan *request.Candles, Max)
	var fe *os.File
	fe, err = os.Open(path)
	if err != nil {
		return err
	}
	defer fe.Close()
	r := bufio.NewReader(fe)
	for {
		db, _, e := r.ReadLine()
		if e != nil {
			break
		}
		self.EndCan = new(request.Candles)
		self.EndCan.Load(string(db))
		self.Can <- self.EndCan
	}
	close(self.Can)
	return nil

}

type Cache struct {
	Scale       int64
	Name        string
	FutureChan  chan *CacheFile
	Stop        chan bool
	EndtimeChan chan int64
	//TmpCan      []*request.Candles
	DiffLong  float64
	LastCan   *request.Candles
	EndJoint  *Joint
	BeginJoin *Joint
	//Id        int
	lastCache *Cache
	par       *InstrumentCache
}

func NewCache(name string, scale int64, p *InstrumentCache) (ca *Cache) {

	ca = &Cache{
		Name:        name,
		Scale:       scale,
		FutureChan:  make(chan *CacheFile, 10),
		Stop:        make(chan bool, 1),
		EndJoint:    new(Joint),
		EndtimeChan: make(chan int64, 10),
		par:         p}
	ca.BeginJoin = ca.EndJoint
	return ca

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
			cf = new(CacheFile)
			er = cf.Init(name, pa, fi, int(86400/self.Scale+1)*2)
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
		beginT, err := time.Parse("2006-01-02T15:04:05", *request.BEGINTIME)
		if err != nil {
			panic(err)
		}
		begin = beginT.Unix()
	} else {
		begin = cf.EndCan.Time + self.Scale
	}
	cf = new(CacheFile)
	cf.Can = make(chan *request.Candles, 1000)
	self.FutureChan <- cf
	request.Down(name, begin, 0, self.Scale, self.Name, func(can *request.Candles) {
		cf.Can <- can
	})
	close(cf.Can)

}

func (self *Cache) CheckUpdate(can *request.Candles) bool {

	if self.LastCan != nil {
		if self.LastCan.Time >= can.Time {
			return false
		}
	}
	self.LastCan = can
	//fmt.Println(can)
	if len(self.par.ServerChan) > 0 {
		for _, ser := range self.par.ServerChan {
			ser.In(TimeCache{
				Scale: self.Scale,
				Time:  can.Time,
				Name:  self.Name})
		}
	}
	return true

}
func (self *Cache) Sensor(cas []*Cache) {
	//var err error
	//var lastCan *request.Candles
	self.Read(func(can *request.Candles) {
		if !self.CheckUpdate(can) {
			return
		}
		endTime := can.Time + self.Scale

		w := new(sync.WaitGroup)
		for _, ca := range cas {
			w.Add(1)
			go func(_ca *Cache) {
				_ca.EndtimeChan <- endTime
				<-_ca.Stop
				w.Done()
			}(ca)
		}
		w.Wait()

		//for _, ca := range cas {
		//	if ca.LastCan != nil {
		//		if ca.LastCan.Time+ca.Scale > endTime {
		//			panic(100)
		//		}
		//	}
		//}

		fmt.Printf("%s\r", time.Unix(endTime, 0))
		//self.CheckTemplagesTest(can, ca_id)

	})

}

func (self *Cache) SyncRun(hand func(can *request.Candles) bool) {
	endTime := <-self.EndtimeChan
	var h func(can *request.Candles)
	h = func(can *request.Candles) {
		if can.Time+self.Scale <= endTime {
			hand(can)
			return
		}
		if len(self.EndtimeChan) == 0 {
			if len(self.Stop) == 0 {
				self.Stop <- true
			}
		}
		endTime = <-self.EndtimeChan
		h(can)
		return
	}
	self.Read(h)
}

func (self *Cache) UpdateJoint(can *request.Candles) bool {

	if !self.CheckUpdate(can) {
		return false
	}
	var up bool
	self.EndJoint, up = self.EndJoint.AppendCans(can)
	if up {
		//endId := len(CacheList) - 1
		if self.lastCache == nil {
			if self.EndJoint.Last != nil && self.EndJoint.Last != self.BeginJoin {
				self.BeginJoin = self.EndJoint.Last
				self.BeginJoin.Last = nil
			}
		} else {
			//endCache := CacheList[endId]
			if len(self.lastCache.BeginJoin.Cans) > 0 {
				endTime := self.lastCache.BeginJoin.Cans[0].Time
				self.BeginJoin.ReadNext(func(jo *Joint) bool {
					if len(jo.Cans) > 0 && jo.Cans[0].Time >= endTime {
						jo.Last = nil
						self.BeginJoin = jo
						return true
					}
					return false
				})
			}
		}
	}

	return true

}

func (self *Cache) Read(Handle func(can *request.Candles)) {

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

//func init() {
//
//	GranularityMap = make(map[string]int64)
//	GranularityMap["S5"] = 5
//	GranularityMap["S10"] = 10
//	GranularityMap["S15"] = 15
//	GranularityMap["S30"] = 30
//	GranularityMap["M1"] = 60
//	GranularityMap["M2"] = 60 * 2
//	GranularityMap["M4"] = 60 * 4
//	GranularityMap["M5"] = 60 * 5
//	GranularityMap["M10"] = 600
//	GranularityMap["M15"] = 60 * 15
//	GranularityMap["M30"] = 60 * 30
//	GranularityMap["H1"] = 3600
//	GranularityMap["H2"] = 3600 * 2
//	GranularityMap["H3"] = 3600 * 3
//	GranularityMap["H4"] = 3600 * 4
//	GranularityMap["H6"] = 3600 * 6
//	GranularityMap["H8"] = 3600 * 8
//	GranularityMap["H12"] = 3600 * 12
//	GranularityMap["D"] = 3600 * 24
//	//GranularityMap["W"] = 3600 * 24 * 7
//	//GranularityMap["M"]= 3600*24*30
//	i := 0
//	for k, v := range GranularityMap {
//		ca := NewCache(k, v)
//		ca.Id = i
//		go ca.Load(filepath.Join(request.Instr.Name, ca.Name))
//		CacheList = append(CacheList, ca)
//		sortCacheList(i)
//		i++
//	}
//	for i, ca := range CacheList {
//		fmt.Println(i, ca.Id)
//	}
//
//}
//func sortCacheList(i int) {
//	if i == 0 {
//		return
//	}
//	I := i - 1
//	if CacheList[I].GetScale() > CacheList[i].GetScale() {
//		CacheList[I], CacheList[i] = CacheList[i], CacheList[I]
//		CacheList[i].Id = i
//		CacheList[I].Id = I
//		sortCacheList(I)
//	}
//
//}
