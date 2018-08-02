package request

import (
	"fmt"
	//"log"
	"os"
	//	"math"
	"io/ioutil"
	"net/http"
	"net/url"
	//	"net"
	//	"golang.org/x/net/proxy"
	"encoding/json"
	//"flag"
	"path/filepath"
	"time"
	"math"
	"strconv"
	"strings"
	//	"bytes"
	"io"
	"github.com/zaddone/RoutineWork/config"
)

var (
	Accounts []*Account
	Client   *http.Client
	Header   http.Header
	//Instr    *Instrument
	ActiveAccount *Account
)

func init() {
	//flag.Parse()
	//Conf = NewConfig()
	Header = make(http.Header)
	Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.Authorization))
	Header.Add("Connection", "Keep-Alive")
	Header.Add("Accept-Datetime-Format", "UNIX")
	Header.Add("Content-type", "application/json")

	if config.Conf.Proxy == "" {
		Client = new(http.Client)
	} else {
		panic(0)
	}
	err := InitAccounts(false)
	if err != nil {
		panic(err)
	}
	SetActiveAccount()

}
func ClientPut(path string, val io.Reader, da interface{}) error {

	Req, err := http.NewRequest("PUT", path, val)
	if err != nil {
		return err
	}
	Req.Header = Header
	res, err := Client.Do(Req)
	if err != nil {
		if res != nil {
			b, err := ioutil.ReadAll(res.Body)
			fmt.Println(string(b), err)
		}
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(da)

}
func ClientPost(path string, val io.Reader, da interface{}) error {

	Req, err := http.NewRequest("POST", path, val)
	if err != nil {
		return err
	}
	Req.Header = Header
	res, err := Client.Do(Req)
	if err != nil {
		if res != nil {
			b, err := ioutil.ReadAll(res.Body)
			fmt.Println(string(b), err)
		}
		return err
	}
	defer res.Body.Close()

	//	b,err:=ioutil.ReadAll(res.Body)
	//	fmt.Println(string(b),err)
	//	return json.Unmarshal(b,da)

	return json.NewDecoder(res.Body).Decode(da)

}

func ClientDOStream(path string) (io.ReadCloser,error) {

	Req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil,err
	}
	Req.Header = Header
	res, err := Client.Do(Req)
	if err != nil {
		return nil,err
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil,fmt.Errorf("status code %d %s", res.StatusCode, path)
	}
	return res.Body,nil

}
func ClientDO(path string, da interface{}) error {
	Req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return err
	}
	Req.Header = Header
	res, err := Client.Do(Req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		//	b,err:=ioutil.ReadAll(res.Body)
		//	fmt.Println(string(b),err)
		return fmt.Errorf("status code %d %s", res.StatusCode, path)
	}
	return json.NewDecoder(res.Body).Decode(da)

	//	b,err:=ioutil.ReadAll(res.Body)
	//	if err != nil {
	//		return err
	//	}
	//	return json.Unmarshal(b,da)

}
func SetActiveAccount() {
	//var Nacc *Account
	//ActiveAccount = nil
	for _, acc := range Accounts {
		if acc.Id == config.Conf.Account_ID {
			ActiveAccount = acc
			ActiveAccount.SetInstruments()
			return
		}
	}
	if ActiveAccount == nil {
		ActiveAccount = Accounts[0]
		err := InitAccounts(true)
		if err != nil {
			panic(err)
		}
		SetActiveAccount()
	}

}

func InitAccounts(update bool) (err error) {
	if !update {
		err = ReadAccounts()
		if err == nil {
			return nil
		}
	}
	da := make(map[string][]*Account)
	err = ClientDO(config.Conf.Host+"/accounts", &da)
	if err != nil {
		return err
	}
	Accounts = da["accounts"]
	//	L := len(self.Accounts)
	if len(Accounts) == 0 {
		fmt.Errorf("accounts == nil")
	}
	return SaveAccounts()

}
func SaveAccounts() error {
	f, err := os.OpenFile(config.Conf.LogFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_SYNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := json.Marshal(Accounts)
	if err != nil {
		return err
	}
	_, err = f.Write(d)
	if err != nil {
		return err
	}
	return nil
}
func ReadAccounts() error {
	fi, err := os.Stat(config.Conf.LogFile)
	if err != nil {
		return err
	}
	data := make([]byte, fi.Size())
	f, err := os.Open(config.Conf.LogFile)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Read(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("%d %d", n, len(data))
	}
	return json.Unmarshal(data, &(Accounts))
}

func Down(name string, from, to int64, gr int64, gran string, Handle func(*Candles)) {

	//gr := GranularityMap[gran]
	var file *os.File = nil
	var LogFile string = ""
	var err error
	var Begin time.Time
	for {
		//		fmt.Println(time.Unix(from,0).UTC(),gran)
		err = GetCandlesHandle(name, gran, from, 500, func(c interface{}) error {
			can := new(Candles)
			can.Init(c.(map[string]interface{}))
			Begin = can.GetTimer()
			Handle(can)
			path := filepath.Join(name, gran, fmt.Sprintf("%d", Begin.Year()))
			_, err := os.Stat(path)
			if err != nil {
				os.MkdirAll(path, 0777)
			}
			path = filepath.Join(path, Begin.Format("20060102"))
			if file == nil {
				LogFile = path
				file, err = os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
				if err != nil {
					panic(err)
				}
			} else if LogFile != path {
				//	fmt.Println(path)
				file.Close()
				LogFile = path
				file, err = os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
				if err != nil {
					panic(err)
				}
			}
			file.WriteString(can.String())
			return nil
		})
		if err != nil {
			//	fmt.Println("err:",err)
			aft := from - time.Now().Unix()
			if aft > 0 {
				<-time.After(time.Second * time.Duration(aft))
			} else {
				<-time.After(time.Second * 3)
			}
		} else {
			from = Begin.Unix() + gr
			if to > 0 && from >= to {
				break
			}
			aft := from - time.Now().Unix()
			if aft > 0 {
				<-time.After(time.Second * time.Duration(aft))
			}
		}
	}

}
func GetCandlesHandle(Ins_name, granularity string, from, count int64, f func(c interface{}) error) (err error) {

	path := fmt.Sprintf("%s/instruments/%s/candles?", config.Conf.Host, Ins_name)
	uv := url.Values{}
	uv.Add("granularity", granularity)
	uv.Add("price", "M")
	uv.Add("from", fmt.Sprintf("%d", from))
	uv.Add("count", fmt.Sprintf("%d", count))
	path += uv.Encode()
	//	fmt.Println(path)
	da := make(map[string]interface{})
	err = ClientDO(path, &da)
	if err != nil {
		return err
	}
	ca := da["candles"].([]interface{})
	lc := len(ca)
	if lc == 0 {
		return fmt.Errorf("candles len = 0")
	}
	//	can = make([]*CandlesMin,lc)
	for _, c := range ca {
		er := f(c)
		if er != nil {
			fmt.Println(er)
			break
		}
		//		can[i] = new(CandlesMin)
		//		can[i].Init(c.(map[string]interface{}))
	}
	return nil

}

type Account struct {
	Id          string                 `json:"id"`
	Tags        []string               `json:"tags"`
	Instruments map[string]*Instrument `json:"instruments"`
}
func (self *Account) GetAccountPathStream() string {
	return fmt.Sprintf("%s/accounts/%s", config.Conf.StreamHost, self.Id)
}
func (self *Account) GetAccountPath() string {
	return fmt.Sprintf("%s/accounts/%s", config.Conf.Host, self.Id)
}
func (self *Account) SetInstruments() error {
	if self.Instruments != nil {
		//fmt.Println(self.Instruments)
		return nil
	}
	path := self.GetAccountPath()
	path += "/instruments"
	da := make(map[string]interface{})
	err := ClientDO(path, &da)
	if err != nil {
		return err
	}
	ins := da["instruments"].([]interface{})
	self.Instruments = make(map[string]*Instrument)
	for _, n := range ins {
		//		fmt.Println(n.(InstrumentTmp).Name)
		in := new(Instrument)
		in.Init(n.(map[string]interface{}))

		self.Instruments[in.Name] = in
	}
	//Nacc.Instruments = ins
	return SaveAccounts()
	//return ins, err

}

//func GetAccountID() string {
//	return Accounts[*Account_ID].Id
//}
type Instrument struct {
	Name string

	DisplayPrecision float64
	MarginRate       float64

	MaximumOrderUnits           float64
	MaximumPositionSize         float64
	MaximumTrailingStopDistance float64

	MinimumTradeSize            float64
	MinimumTrailingStopDistance float64

	PipLocation         float64
	TradeUnitsPrecision float64
	Type                string
	Online              bool
	pipDiff  float64
}
func (self *Instrument) PipDiff() float64 {
	if self.pipDiff == 0 {
		return math.Pow10(-int(self.DisplayPrecision))*(-self.PipLocation)*2
	}else{
		return self.pipDiff
	}
}
func (self *Instrument) StandardPrice(pr float64) string {

	_int,frac:= math.Modf(pr)
	frac  = math.Pow10(int(self.DisplayPrecision)) * frac
	return fmt.Sprintf("%d.%d", int(_int),int(frac))

}

func (self *Instrument) Init(tmp map[string]interface{}) (err error) {
	self.Name = tmp["name"].(string)
	self.PipLocation = tmp["pipLocation"].(float64)
	self.TradeUnitsPrecision = tmp["tradeUnitsPrecision"].(float64)
	self.Type = tmp["type"].(string)
	self.DisplayPrecision = tmp["displayPrecision"].(float64)

	self.MarginRate, err = strconv.ParseFloat(tmp["marginRate"].(string), 64)
	if err != nil {
		return err
	}
	self.MaximumOrderUnits, err = strconv.ParseFloat(tmp["maximumOrderUnits"].(string), 64)
	if err != nil {
		return err
	}
	self.MaximumPositionSize, err = strconv.ParseFloat(tmp["maximumPositionSize"].(string), 64)
	if err != nil {
		return err
	}
	self.MaximumTrailingStopDistance, err = strconv.ParseFloat(tmp["maximumTrailingStopDistance"].(string), 64)
	if err != nil {
		return err
	}
	self.MinimumTradeSize, err = strconv.ParseFloat(tmp["minimumTradeSize"].(string), 64)
	if err != nil {
		return err
	}
	self.MinimumTrailingStopDistance, err = strconv.ParseFloat(tmp["minimumTrailingStopDistance"].(string), 64)
	if err != nil {
		return err
	}

	return nil
}

type Candles struct {
	Mid    [4]float64
	Time   int64
	Volume float64
	Val    float64
}

func (self *Candles) GetMidLong() float64 {
	return self.Mid[2] - self.Mid[3]
}

func (self *Candles) GetMidAverage() float64 {
	if self.Val == 0 {
		var sum float64 = 0
		for _, m := range self.Mid {
			sum += m
		}
		self.Val = sum / 4
	}
	return self.Val
}

func (self *Candles) GetInOut() bool {
	return self.Mid[0] < self.Mid[1]
}

func (self *Candles) GetTime() int64 {
	return self.Time
}
func (self *Candles) GetTimer() time.Time {
	return time.Unix(self.Time, 0).UTC()

}

func (self *Candles) Show() {
	fmt.Printf("%.6f %.6f %.6f %.6f %s %.6f\r\n", self.Mid[0], self.Mid[1], self.Mid[2], self.Mid[3], time.Unix(self.Time, 0).String(), self.Volume)
}

func (self *Candles) Load(str string) (err error) {
	strl := strings.Split(str, " ")
	self.Mid[0], err = strconv.ParseFloat(strl[0], 64)
	if err != nil {
		return err
	}
	self.Mid[1], err = strconv.ParseFloat(strl[1], 64)
	if err != nil {
		return err
	}
	self.Mid[2], err = strconv.ParseFloat(strl[2], 64)
	if err != nil {
		return err
	}
	self.Mid[3], err = strconv.ParseFloat(strl[3], 64)
	if err != nil {
		return err
	}
	Ti, err := strconv.Atoi(strl[4])
	if err != nil {
		return err
	}
	self.Time = int64(Ti)
	self.Volume, err = strconv.ParseFloat(strl[5], 64)
	if err != nil {
		return err
	}
	return nil
}

func (self *Candles) String() string {
	return fmt.Sprintf("%.5f %.5f %.5f %.5f %d %.5f\r\n", self.Mid[0], self.Mid[1], self.Mid[2], self.Mid[3], self.Time, self.Volume)
}
func (self *Candles) CheckSection( val float64 ) bool {
	fmt.Println(self.Mid[2] , val , self.Mid[3])
	if val > self.Mid[2] || val < self.Mid[3] {
		return false
	}
	return true
}
func (self *Candles) Init(tmp map[string]interface{}) (err error) {
	Mid := tmp["mid"].(map[string]interface{})
	if Mid != nil {
		self.Mid[0], err = strconv.ParseFloat(Mid["o"].(string), 64)
		if err != nil {
			return err
		}
		self.Mid[1], err = strconv.ParseFloat(Mid["c"].(string), 64)
		if err != nil {
			return err
		}
		self.Mid[2], err = strconv.ParseFloat(Mid["h"].(string), 64)
		if err != nil {
			return err
		}
		self.Mid[3], err = strconv.ParseFloat(Mid["l"].(string), 64)
		if err != nil {
			return err
		}
	}
	self.Volume = tmp["volume"].(float64)
	ti, err := strconv.ParseFloat(tmp["time"].(string), 64)
	if err != nil {
		return err
	}
	self.Time = int64(ti)
	return nil
}
