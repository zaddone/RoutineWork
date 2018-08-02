package replay
import (
	"os"
	"bufio"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
)
type CacheFile struct {
	Can    chan config.Candles
	Fi     os.FileInfo
	Path   string
	EndCan config.Candles
}
func NewCacheFile(name, path string, fi os.FileInfo, Max int) (*CacheFile,error) {

	var cf CacheFile
	return &cf,cf.Init(name,path,fi,Max)

}

func (self *CacheFile) Init(name, path string, fi os.FileInfo, Max int) (err error) {

	self.Path = path
	self.Fi = fi
	self.Can = make(chan config.Candles, Max)
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


