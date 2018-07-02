package server

import (
	pb "github.com/zaddone/RoutineWork/console"
	"github.com/zaddone/RoutineWork/replay"
	"github.com/zaddone/RoutineWork/request"
	"github.com/zaddone/RoutineWork/config"
	"context"
	//"flag"
	"time"
	//"fmt"
	//"golang.org/x/net/context"
	//"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

var (
	//port = flag.String("p", ":50051", "Input Port")
)

type server struct{}

func (s *server) StartInstrument(ctx context.Context, InsName *pb.InstrumentSimple) (*pb.InstrumentReply, error) {

	return &pb.InstrumentReply{State: replay.Start(InsName.Name)}, nil

}
func (s *server) GetLastTime(InsName *pb.InstrumentSimple, stream pb.Greeter_GetLastTimeServer) error {
	for _, inc := range replay.InsCaches {
		if inc.Ins.Name == InsName.Name {
			log.Println(InsName, "lastTime")

			ser := new(replay.ServerChan)
			ctx, cancel := context.WithCancel(context.Background())
			ser.Init(ctx)
			key := int(time.Now().UnixNano())
			inc.ServerChanMap.Add(key,ser)
			defer func() {
				log.Println("lastTime over")
				inc.ServerChanMap.Del(key)
				cancel()

			}()
			return ser.Out(func(ti *replay.TimeCache) error {

				return stream.Send(&pb.LastTime{
					Tag:   ti.Name,
					Scale: ti.Scale,
					Time:  ti.Time})

			})

		}
	}
	return nil

}

func (s *server) ListInstrument(re *pb.Request, stream pb.Greeter_ListInstrumentServer) error {

	instr := request.ActiveAccount.Instruments
	for _, inc := range replay.InsCaches {
		instr[inc.Ins.Name].Online = true
	}

	for _, ins := range instr {
		if err := stream.Send(
			&pb.Instrument{
				Name:                        ins.Name,
				DisplayPrecision:            ins.DisplayPrecision,
				MarginRate:                  ins.MarginRate,
				MaximumOrderUnits:           ins.MaximumOrderUnits,
				MaximumPositionSize:         ins.MaximumPositionSize,
				MaximumTrailingStopDistance: ins.MaximumTrailingStopDistance,
				MinimumTradeSize:            ins.MinimumTradeSize,
				MinimumTrailingStopDistance: ins.MinimumTrailingStopDistance,
				PipLocation:                 ins.PipLocation,
				TradeUnitsPrecision:         ins.TradeUnitsPrecision,
				Online:                      ins.Online,
				Type:                        ins.Type}); err != nil {
			return err
		}

	}

	return nil
}

func init() {
	//flag.Parse()
	if !config.Conf.Server {
		return
	}
	lis, err := net.Listen("tcp", config.Conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to server: %v", err)
		}
	}()
	log.Println("server run",config.Conf.Port)
}
