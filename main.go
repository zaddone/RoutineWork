package main

import (
	_ "github.com/zaddone/RoutineWork/replay"
	_ "github.com/zaddone/RoutineWork/request"
	_ "github.com/zaddone/RoutineWork/server"
	"fmt"
)

func main() {

	//InsCache := new(replay.InstrumentCache)
	//InsCache.Init(request.Instr.Name)
	//InsCache.Run()
	//	for _, _ca := range replay.CacheList[1:] {
	//		go _ca.SyncRun(_ca.UpdateJoint)
	//	}
	//	replay.CacheList[0].Sensor(replay.CacheList[1:])

	var cmd string
	for {
		fmt.Scanf("%s\r", &cmd)
		fmt.Println(cmd)
	}

}
