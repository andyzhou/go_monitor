package monitor

import (
	"fmt"
	"github.com/andyzhou/monitor/define"
	"github.com/andyzhou/monitor/face"
	"github.com/andyzhou/monitor/service"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
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
	//logService := tc.NewLogService(Options.logPath, Options.logPrefix)

	//catch signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	//signal snatch
	go func(wg *sync.WaitGroup) {
		for {
			select {
			case s := <- c:
				{
					log.Println("get signal of ", s.String())
					wg.Done()
					return
				}
			}
		}
	}(&wg)

	//init inter face
	face.RunInterFace = face.NewInterFace()

	//start wait group
	wg.Add(1)

	//set port
	port := Options.port
	if port <= 0 {
		port = define.DefaultPort
	}

	//start rpc service
	fmt.Printf("start monitor on %v...\n", port)
	rpcService := service.NewRpcServer(port)

	wg.Wait()

	//clean service and data
	rpcService.Stop()
	face.RunInterFace.Quit()
	//logService.Close()
}
