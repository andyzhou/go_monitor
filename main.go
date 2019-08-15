package main

import (
	"bitbucket.org/arui/tinycells/tc"
	"monitor/service"
	"sync"
	"log"
	"runtime/debug"
	"os"
	"os/signal"
	"monitor/face"
)

func init()  {
	parseFlags()
}

func main()  {
	var wg sync.WaitGroup

	//try catch main panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Panic happened, error:", err)
			log.Println("Stack:", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	//pre init log service
	logService := tc.NewLogService(Options.logPath, Options.logPrefix)

	//catch signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	//signal snatch
	go func(wg *sync.WaitGroup) {
		var needQuit bool
		for {
			if needQuit {
				break
			}
			select {
			case s := <- c:
				log.Println("get signal of ", s.String())
				wg.Done()
				needQuit = true
				break
			}
		}
	}(&wg)

	//init inter face
	face.RunInterFace = face.NewInterFace()

	//start wait group
	wg.Add(1)

	//start rpc service
	rpcService := service.NewRpcServer(Options.port)

	wg.Wait()

	//clean service and data
	rpcService.Stop()
	face.RunInterFace.Quit()
	logService.Close()
}
