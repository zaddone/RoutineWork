package signal
import(
	"fmt"
	"math"
	"github.com/zaddone/RoutineWork/config"
)

type Joint struct{
	Cans    []config.Candles
	Last    *Joint
	Next    *Joint
	//Dur     *bool
	SumLong float64
	Diff    float64
	//BDiff   float64
	MaxDiff float64
	Num	int
	Scale   int64
}

func NewJoint(cans []config.Candles) *Joint {
	var jo Joint
	le := len(cans)
	if le > 0 {
		jo.Cans = make([]config.Candles, le)
		jo.SumLong = 0
		var lastCa config.Candles
		var scale int64
		for i, _ca := range cans {
			jo.SumLong += _ca.GetMidLong()
			jo.Cans[i] = _ca
			if lastCa == nil {
				lastCa = _ca
			}else{
				scale = _ca.GetTime()- lastCa.GetTime()
				if jo.Scale ==0 || jo.Scale > scale {
					jo.Scale = scale
				}
			}
		}
		jo.Diff = cans[le-1].GetMidAverage() - cans[0].GetMidAverage()
	}
	//jo.BDiff = jo.Diff
	return &jo

}

func (self *Joint) ratioTime() float64 {

	if self.Diff == 0 {
		fmt.Println(len(self.Cans))
		panic("diff == 0")
	}
	le := len(self.Cans)
	if le == 0 {
		panic("le == 0")
	}
	time := (self.Cans[le-1].GetTime() - self.Cans[0].GetTime())/self.Scale +1
	if time == 0 {
		panic("time == 0")
	}
	return float64(time)/math.Abs(self.Diff)

}

func (self *Joint) ratioAve() float64 {

	//fmt.Println(len(self.Cans),self.GetLongAve(),self.Diff)
	if self.Diff == 0 {
		panic("diff == 0")
	}
	ave := self.GetLongAve()
	if ave == 0 {
		panic("ave == 0")
	}
	return ave/math.Abs(self.Diff)

}

func (self *Joint) ratio() (x float64, y float64) {
	return self.ratioTime(),self.ratioAve()
}

func (self *Joint) ReadLast(hand func(jo *Joint) bool ) {
	if hand(self) {
		return
	}
	if self.Last != nil {
		self.Last.ReadLast(hand)
	}
	return
}

func (self *Joint) ReadNextCan (begin int,f func(can config.Candles) bool) {
	le:= len(self.Cans)
	for i:=begin;i<le;i++ {
		if f(self.Cans[i]) {
			return
		}
	}
	if self.Next != nil {
		self.Next.ReadNextCan(0,f)
	}
}
func (self *Joint) Find (can config.Candles) (*Joint,int) {

	le := len(self.Cans)
	if le == 0 {
		if self.Next == nil {
			return self,-1
			//panic("self.Next == nil")
		}
		return self.Next.Find(can)
	}
	if can.GetTime() < self.Cans[0].GetTime() {
		if self.Last == nil {
			return self,-1
		}
		return self.Last.Find(can)
		//panic("c < 0")
	}
	end:= le-1
	if can.GetTime() > self.Cans[end].GetTime() {
		return self.Next.Find(can)
	}
	return self,self.binaryChop(0,end,can)

}
func (self *Joint) binaryChop(b,e int,can config.Candles ) int{
	f := e-b
	if f<0 {
		return f
	}
	index := b + f/2
	_time :=self.Cans[index].GetTime()
	ctime := can.GetTime()
	if _time > ctime {
		return self.binaryChop(b,index-1,can)
	}else if _time < ctime {
		return self.binaryChop(index+1,e,can)
	}else{
		if can!= self.Cans[index] {
			panic("is not same")
		}
		return index
	}
}

func (self *Joint) Reload(cans []config.Candles) {

	le := len(cans)
	self.Cans = make([]config.Candles, le)
	self.SumLong = 0
	for i, _ca := range cans {
		self.SumLong += _ca.GetMidLong()
		self.Cans[i] = _ca
	}
	self.Diff = cans[le-1].GetMidAverage() - cans[0].GetMidAverage()
	//self.Dur = nil

}

func (self *Joint) split(id int) (jo *Joint) {

	jo = NewJoint(self.Cans[id:])
	jo.Last = self

	self.Reload(self.Cans[:id])
	self.Next = jo
	self.Diff = jo.Cans[0].GetMidAverage() - self.Cans[0].GetMidAverage()
	jo.Num = self.Num+1
	if (jo.Diff>0) == (self.Diff>0) {
		panic("jo f diff")
	}
	//fmt.Println(self.Scale,len(self.Cans))
	return jo

}
func (self *Joint) Check(){

	if self.Last == nil {
		return
	}
	f := self.Diff>0
	var Var,Var1 Variance
	n:=0
	fmt.Println("b__________")
	self.Last.ReadLast(func (jo *Joint) bool {
		n++

		if f == (jo.Diff>0) {
			Var.add(jo.ratioAve())
		}else{
			Var1.add(jo.ratioAve())
		}
		fmt.Println(Var.GetAve(),Var.GetVal(),Var1.GetAve(),Var1.GetVal())
		return false

	})
	fmt.Println("e__________")

}

func (self *Joint) merges(jo *Joint) {

	self.Cans = append(self.Cans,jo.Cans...)
	jo.Cans = nil
	self.SumLong += jo.SumLong
	self.MaxDiff = 0

}

func (self *Joint) merge() {

	LastC := self.Last
	if LastC == nil {
		return
	}
	//LastC.Next = nil
	self.Last = LastC.Last
	if self.Last != nil {
		//LastC.Last = nil
		self.Last.Next = self
	}
	self.Cans = append(LastC.Cans, self.Cans...)
	LastC.Cans = nil
	self.SumLong += LastC.SumLong
	self.Num = LastC.Num

}
func (self *Joint) Append(can config.Candles,scale int64) (jo *Joint, update bool) {
	canVal := can.GetMidAverage()
	self.Cans = append(self.Cans, can)
	self.Scale = scale
	self.SumLong += can.GetMidLong()
	self.Diff = canVal - self.Cans[0].GetMidAverage()
	jo = self
	if self.Last != nil {
		if (self.Last.Diff>0) == (self.Diff>0) {
			jo = self.Last
			jo.merges(self)
			return
		}
	}
	le := len(self.Cans)
	le--
	if le < 3 {
		return
	}
	self.MaxDiff = 0
	ave := self.GetLongAve()
	var dif float64 = 0
	var maxId int = -1
	for i := 1; i < le; i++ {
		dif = canVal - self.Cans[i].GetMidAverage()
		if (dif > 0) != (self.Diff > 0) {
			dif = math.Abs(dif)
			if (dif > self.MaxDiff) {
				self.MaxDiff = dif
				maxId = i
			}
		}
	}
	if ave > self.MaxDiff {
		return
	}
	//fmt.Println("split",maxId,le,ave , self.MaxDiff)
	//if maxId < 3 {
	//	return
	//}
	jo = self.split(maxId)
	//fmt.Println("jo",len(jo.Cans))
	update = true
	return

}
func (self *Joint) ReadSameDiff(f func(*Joint) bool){

	if self.Last == nil {
		return
	}
	laJo :=self.Last.Last
	if  laJo == nil {
		return
	}
	if f(laJo) {
		return
	}
	laJo.ReadSameDiff(f)

}

func (self *Joint) GetLongAve() float64 {
	if self.SumLong == 0 {
		for _,can := range self.Cans {
			self.SumLong += can.GetMidLong()
		}
	}
	return self.SumLong / float64(len(self.Cans))
}

func (self *Joint) Cut() {
	if self.Last != nil {
		self.Last.Next = nil
		self.Last = nil
	}
}
