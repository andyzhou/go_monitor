package main

import (
	"log"
	"os"
	"monitor/client"
	"sync"
	"strconv"
)

//internal macro variables
const (
	RpcHost = "127.0.0.1"
	RpcPort = 9100
)

//call back for node changed notify
func notify(node client.NodeInfo) bool {
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
		log.Println("Useage:", os.Args[0], " <host> <port>")
		os.Exit(1)
	}

	host = os.Args[1]
	port, _ = strconv.Atoi(os.Args[2])

	//init and add new monitor
	client := client.NewMonitorClient()

	//set client node
	client.SetClient("chat", host, int32(port))

	//set call back func
	client.SetNotifyFunc(notify)

	//try ping monitors?
	client.AddMonitor(RpcHost, RpcPort)

	result := client.GetBatchNodes("")
	log.Println("result:", result)

	wg.Add(1)
	wg.Wait()

	client.CleanUp()
}

