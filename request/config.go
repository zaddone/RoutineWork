package request
import(
	//"flag"
	"github.com/BurntSushi/toml"
	"os"
	"log"
)
const (
	FileName string = "request.log"
)

type Config struct {
	Account_ID string
	Authorization string
	Proxy string
	LogFile string
	Host string
	BEGINTIME string
	Port string
}
func (self *Config) Save() {

	fi,err := os.OpenFile(FileName,os.O_CREATE|os.O_WRONLY,0777)
	//fi,err := os.Open(FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()
	e := toml.NewEncoder(fi)
	err = e.Encode(self)
	if err != nil {
		log.Fatal(err)
	}

}
func NewConfig()  *Config {

	var c Config
	_,err := os.Stat(FileName)
	if err != nil {
		c.Account_ID = "101-011-2471429-001"
		c.Authorization = ""
		c.Proxy = ""
		c.LogFile = "LogInfo.log"
		c.Host = "https://api-fxpractice.oanda.com/v3"
		c.BEGINTIME = "2009-01-01T00:00:00"
		c.Port=":50051"
		c.Save()
	}else{
		if _,err := toml.DecodeFile(FileName,&c);err != nil {
			log.Fatal(err)
		}
	}
	return &c

}
