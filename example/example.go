package main

import (
	"github.com/andyzhou/monitor"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

/*
 * example code
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro variables
const (
	RpcHost = "127.0.0.1"
	RpcPort = 9100
)

//call back for node changed notify
func notify(node monitor.NodeInfo) bool {
	log.Println("node changed, node:", node)
	return true
}

func main()  {
	var (
		wg sync.WaitGroup
		host string
		port int
		args = len(os.Args)
	)

	if args < 3 {
		log.Println("Usage:", os.Args[0], " <host> <port>")
		os.Exit(1)
	}

	host = os.Args[1]
	port, _ = strconv.Atoi(os.Args[2])

	//init and add new monitor
	client := monitor.NewMonitorClient()

	//set client node
	client.SetClient("chat", host, int32(port))

	//set call back func
	client.SetNotifyFunc(notify)

	//try ping monitors?
	client.AddMonitor(RpcHost, RpcPort)

	//get batch nodes
	sf := func() {
		for {
			result := client.GetBatchNodes("")
			log.Println("result:", result)
			time.Sleep(time.Second * 3)
		}
	}
	go sf()

	wg.Add(1)
	wg.Wait()

	client.CleanUp()
}

