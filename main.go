package main

import (
	_ "github.com/zaddone/RoutineWork/replay"
	_ "github.com/zaddone/RoutineWork/request"
	_ "github.com/zaddone/RoutineWork/server"
	"fmt"
)

func main() {

	var cmd string
	for {
		fmt.Scanf("%s\r", &cmd)
		fmt.Println(cmd)
	}

}
