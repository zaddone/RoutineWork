package signal
import(
	"fmt"
	"math"
)

type Cov struct {

	jos []*Joint
	xSum float64
	ySum float64
	xySum float64
	num float64
	val float64

}
func (self *Cov) GetYAve() float64 {
	return self.ySum/self.num
}
func (self *Cov) GetXAve() float64 {
	return self.xSum/self.num
}
func (self *Cov) AppendJos(jo *Joint) {

	self.jos = append(self.jos,jo)

}
func (self *Cov) GetJoints() []*Joint {

	return self.jos

}
func (self *Cov) Dels(index int){
	var x,y float64
	for _,jo := range self.jos[:index] {
		if len(jo.Cans) == 0 {
			continue
		}
		x,y = jo.ratio()
		self.xSum -= x
		self.ySum -= y
		self.xySum-= x*y
		self.num --
	}
	self.jos = self.jos[index:]
	self.SetVal()
}
func (self *Cov) Load(jos []*Joint){
	self.jos = jos
	self.xSum = 0
	self.ySum = 0
	self.xySum = 0
	self.num = 0
	var x,y float64
	for _,jo := range self.jos {

		x,y = jo.ratio()
		self.xSum += x
		self.ySum += y
		self.xySum+= x*y
		self.num ++
	}
	self.SetVal()
}
func (self *Cov) GetVal() float64 {
	return self.val
}
func (self *Cov) SetVal(){
	self.val = self.xySum/self.num - (self.xSum*self.ySum)/(self.num*self.num)
	if math.IsInf(self.val,1) || math.IsNaN(self.val) {
		panic("val is Inf or is NaN")
	}
}
func (self *Cov) Del() (jo *Joint){

	jo = self.jos[0]
	self.jos = self.jos[1:]
	if len(jo.Cans) == 0 {
		return
	}
	x,y :=jo.ratio()
	self.xSum -= x
	self.ySum -= y
	self.xySum -= x*y
	self.num --
	self.SetVal()
	return

}
func (self *Cov) GetCount() int{
	return len(self.jos)
}
func (self Cov) DummyAdd (jo *Joint) *Cov {

	x,y :=jo.ratio()
	self.xSum += x
	if math.IsInf(self.xSum,1) || math.IsNaN(self.xSum) {
		panic("x is Inf or is NaN")
	}
	self.ySum += y
	if math.IsInf(self.ySum,1) || math.IsNaN(self.ySum) {
		panic("y is Inf or is NaN")
	}
	self.xySum+= x*y
	if math.IsInf(self.xySum,1) || math.IsNaN(self.xySum) {
		panic("xy is Inf or is NaN")
	}
	self.num ++
	self.SetVal()
	if math.IsInf(self.val,1) ||
	 math.IsNaN(self.val) {
		panic("val is Inf or is NaN")
	}
	self.jos = append(self.jos,jo)
	return &self

}
type Variance struct {

	jos []*Joint
	sum float64
	squareSum float64
	num float64
	val float64

}
func (self Variance) DummyAdd (jo *Joint) *Variance {

	v :=jo.ratioAve()
	self.sum+=v
	if math.IsInf(self.sum,1) {
		return nil
	}
	self.squareSum += math.Pow(v,2)
	if math.IsInf(self.squareSum,1) {
		return nil
	}
	self.num ++
	self.SetVal()
	if math.IsInf(self.val,1) ||
	 math.IsNaN(self.val) {
		return nil
	}
	self.jos = append(self.jos,jo)
	return &self

}
func (self *Variance) GetJoints() []*Joint {

	return self.jos

}
func (self *Variance) Load(jos []*Joint){
	self.jos = jos
	self.sum = 0
	self.squareSum = 0
	self.num = 0
	var v float64
	for _,jo := range self.jos {
		v = jo.ratioAve()
		self.sum += v
		self.squareSum+=math.Pow(v,2)
		self.num++
	}
	self.SetVal()
}
func (self *Variance) DummyDel() (jo *Joint) {

	jo = self.jos[0]
	self.jos = self.jos[1:]
	v := jo.ratioAve()
	self.sum -= v
	self.squareSum-=math.Pow(v,2)
	self.num --
	self.SetVal()
	return

}
func (self *Variance) SetVal(){
	self.val = math.Sqrt((self.squareSum/self.num) - ( math.Pow(self.sum,2) / math.Pow(self.num,2) ))
}

func (self *Variance) GetCount() int {
	return len(self.jos)
}
func (self *Variance) add (v float64) bool {
	sum := self.sum + v
	if math.IsInf(sum,1) {
		return false
	}
	squareSum:= self.squareSum + math.Pow(v,2)
	if math.IsInf(squareSum,1) {
		return false
	}
	num := self.num+1
	val :=math.Sqrt((squareSum/num) - ( math.Pow(sum,2) / math.Pow(num,2) ))
	if math.IsInf(val,1) {
		return false
	}
	if math.IsNaN(val) {
		fmt.Printf("%.5f %.5f %.5f %.0f \r\n",v,squareSum,sum,num)
		return false
	}

	self.squareSum = squareSum
	self.num = num
	self.sum = sum
	self.val = val
	//fmt.Println(v,self.sum,self.squareSum,val,"--------------")
	return true

}
func (self *Variance) GetAve() float64 {
	return self.sum/self.num
}
//func (self *Variance) del(v float64) {
//
//	self.sum -= v
//	self.squareSum -= math.Pow(v,2)
//	self.num --
//	fmt.Println(self.sum,self.squareSum,self.num,"---------------")
//
//}
func (self *Variance) GetVal() float64 {
	return self.val
}
