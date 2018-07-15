package replay

import (
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"bufio"
	"fmt"
	//"math"
	"os"
	"path/filepath"
	//"strconv"
	"time"
	//"strings"
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
func NewCacheFile(name, path string, fi os.FileInfo, Max int) (*CacheFile,error) {

	var cf CacheFile
	err := cf.Init(name,path,fi,Max)
	return &cf,err

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
	//Stop        chan bool
	//w           *sync.WaitGroup
	EndtimeChan chan int64
	//TmpCan      []*request.Candles
	DiffLong  float64
	LastCan   *request.Candles

	EndJoint  *Joint
	BeginJoint *Joint

	//JointLib  JointLib
	//Id        int
	LastCache *Cache
	par       *InstrumentCache
	IsSplit bool
	//IsUpdate bool
}
func (self *Cache) GetInsCache() *InstrumentCache {
	return self.par
}
func (self *Cache) GetMinCache() *Cache {
	return self.par.GetBaseCache()
}
func (self *Cache) GetMinCan() *request.Candles {
	return self.par.GetBaseCan()
}
func (self *Cache) GetInsName() string {
	return self.par.Ins.Name
}

func NewCache(name string, scale int64, p *InstrumentCache) (ca *Cache) {

	ca = &Cache{
		Name:        name,
		Scale:       scale,
		FutureChan:  make(chan *CacheFile, 10),
		//Stop:        make(chan bool, 1),
		EndJoint:    new(Joint),
		EndtimeChan: make(chan int64, 1),
		par:         p}
	ca.BeginJoint = ca.EndJoint
	//ca.JointLib.New()
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
		begin = cf.EndCan.Time + self.Scale
	}
	cf = &CacheFile{Can:make(chan *request.Candles, 1000)}
	//cf.Can = make(chan *request.Candles, 1000)
	self.FutureChan <- cf
	request.Down(name, begin, 0, self.Scale, self.Name, func(can *request.Candles) {
		cf.Can <- can
	})
	close(cf.Can)

}

func (self *Cache) CheckUpdate(can *request.Candles) bool {

	if can.GetMidLong() == 0 {
		return false
	}
	if self.LastCan != nil {
		if self.LastCan.Time >= can.Time {
			return false
		}
	}
	self.LastCan = can
	//fmt.Println(can)
	return true

}
func (self *Cache) Sensor(cas []*Cache) {
	calen := len(cas)
	//self.IsUpdate = false
	var date string = ""
	self.Read(func(can *request.Candles) {
		if !self.UpdateJoint(can) {
			return
		}
		endTime := can.Time + self.Scale
		self.par.w.Add(calen)
		for _, ca := range cas {
			go func(_ca *Cache) {
				_ca.IsSplit = false
				_ca.EndtimeChan <- endTime
			}(ca)
		}
		self.par.w.Wait()

		day:= time.Unix(endTime, 0)
		//d := day.Format("2006-01-02")
		d := day.Format("2006-01")
		if d != date {
			SignalBox[3] = SignalBox[1]/SignalBox[0]
			fmt.Printf("%s %.5f---------------------\r\n",date,SignalBox)
			date = d
			SignalBox=[6]float64{0,0,0,0,0,0}
		}
		//if self.par.Price.Time != "" {
		//	ti := strings.Split(string(self.par.Price.Time),".")
		//	sec,err := strconv.Atoi(ti[0])
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//	nsec,err := strconv.Atoi(ti[1])
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//	for i,b := range self.par.Price.Bids {
		//		fmt.Println(i,b.Price.GetPrice(),self.par.Price.CloseoutBid)
		//	}
		//	fmt.Printf("%s %.2f %s %d %d\r", day , SignalBox[:3],self.par.Price.Bids,int64(sec),int64(nsec))
		//}else{
		//fmt.Printf("%s %.2f\r", day , SignalBox)
		//}
		self.par.Signal()
		//if cas[0].LastCache != nil {
		//	fmt.Printf("%s %s\r", time.Unix(endTime, 0),time.Unix(cas[0].LastCan.Time,0))
		//}else{
		//}
	})
}

func (self *Cache) SyncRun(hand func(can *request.Candles) bool) {
	self.Read(func(can *request.Candles){
		//self.IsUpdate = false
		for{
			endTime := <-self.EndtimeChan
			if can.Time+self.Scale <= endTime {
				hand(can)
				//fmt.Println(self.Name)
				//self.IsUpdate = true
				//self.par.w.Done()
				self.EndtimeChan <- endTime
				return
			}
			self.par.w.Done()
		}
	})
}

func (self *Cache) UpdateJoint(can *request.Candles) ( bool) {

	if !self.CheckUpdate(can) {
		return false
	}
	self.par.Monitor(self,can)
	//self.JointLib.Update(can)

	self.EndJoint, self.IsSplit = self.EndJoint.Append(can)
	//fmt.Println(len(self.par.SplitCache))
	if self.IsSplit {
		if self.LastCache == nil {
			j :=0
			self.EndJoint.ReadLast(func(jo *Joint) bool {
				j++
				if j < 4 {
					return false
				}
				jo.Cut()
				self.BeginJoint = jo
				//fmt.Println("end cache",self.Name,time.Unix(jo.Cans[0].Time,0))
				return true
			})
		} else {
			//self.par.SplitCache <- self
			if len(self.LastCache.BeginJoint.Cans) > 0 {
				endTime := self.LastCache.BeginJoint.Cans[0].Time
				self.BeginJoint.ReadNext(func(jo *Joint) bool {
					if len(jo.Cans) > 0 && jo.Cans[0].Time < endTime {
						return false
					}
					if self.BeginJoint != jo {
						self.BeginJoint = jo
						jo.Cut()
						//fmt.Println(self.Name,self.BeginJoint.Cans[0].Time,endTime)
					}
					return true
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
