package signal
import(
	"github.com/zaddone/RoutineWork/replay"
)

func init(){
	replay.SignalGroup = append(replay.SignalGroup, NewAna())
}
