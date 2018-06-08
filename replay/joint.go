package replay

import (
	"github.com/zaddone/RoutineWork/request"
	"math"
)

type Joint struct {
	Cans    []*request.Candles
	Last    *Joint
	Next    *Joint
	Dur     *bool
	SumLong float64
	Diff    float64
}

func NewJoint(cans []*request.Candles) *Joint {
	var jo Joint
	le := len(cans)
	jo.Cans = make([]*request.Candles, le)
	jo.SumLong = 0
	for i, _ca := range cans {
		jo.SumLong += _ca.GetMidLong()
		jo.Cans[i] = _ca
	}
	jo.Diff = cans[le-1].GetMidAverage() - cans[0].GetMidAverage()
	return &jo
}

func (self *Joint) ReadNext(hand func(jo *Joint) bool) {

	if hand(self) {
		return
	}
	if self.Next != nil {
		self.Next.ReadNext(hand)
	}
	return

}

func (self *Joint) Reload(cans []*request.Candles) {
	le := len(cans)
	self.Cans = make([]*request.Candles, le)
	self.SumLong = 0
	for i, _ca := range cans {
		self.SumLong += _ca.GetMidLong()
		self.Cans[i] = _ca
	}
	self.Diff = cans[le-1].GetMidAverage() - cans[0].GetMidAverage()
	self.Dur = nil

}

func (self *Joint) split(id int) (jo *Joint) {

	jo = NewJoint(self.Cans[id:])
	jo.Last = self
	self.Reload(self.Cans[:id])
	self.Next = jo
	return jo

}

func (self *Joint) merge() {
	if self.Last == nil {
		return
	}
	self.Cans = append(self.Last.Cans, self.Cans...)
	self.SumLong += self.Last.SumLong
	self.Last = self.Last.Last
	if self.Last != nil {
		self.Last.Next = self
	}

}

func (self *Joint) AppendCans(can *request.Candles) (jo *Joint, update bool) {
	jo = self
	update = false
	canVal := can.GetMidAverage()
	var tmpDiff float64
	defer func() {
		jo.Cans = append(jo.Cans, can)
		jo.SumLong += can.GetMidLong()
		jo.Diff = tmpDiff
	}()
	le := len(self.Cans)
	if le < 2 {
		return
	}

	tmpDiff = canVal - jo.Cans[0].GetMidAverage()
	if (tmpDiff > 0) != (self.Diff > 0) {
		jo.merge()
		return
	}

	ave := self.GetLongAve()
	var dif, maxDif float64 = 0, 0
	var maxId int = 0
	for i := 1; i < le; i++ {
		dif = canVal - self.Cans[i].GetMidAverage()
		if (dif > 0) != (self.Diff > 0) {
			dif = math.Abs(dif)
			if (dif > ave) && (dif > maxDif) {
				maxDif = dif
				maxId = i
			}
		}
	}
	if maxId == 0 {
		return
	}
	if ave > math.Abs(maxDif) {
		return
	}
	jo = self.split(maxId)
	update = true
	return

}
func (self *Joint) GetLongAve() float64 {
	return self.SumLong / float64(len(self.Cans))
}
