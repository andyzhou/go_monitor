package client

import (
	"github.com/andyzhou/tinycells/tc"
	"google.golang.org/grpc"
	pb "github.com/andyzhou/monitor/pb"
	"sync"
	"fmt"
	"log"
	"context"
	"time"
	"io"
)

/*
 * monitor client API
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * support multi monitors?
 */

 //define node status
 const (
 	NodeStatusDown = iota //node down
 	NodeStatusUp //node up
 	NodeStatusMaintain //node in maintain
 )

 //internal macro variables
 const (
 	MonitorCheckRate = 5 //xxx seconds
 )

 //single node info
 type NodeInfo struct {
 	Kind string `node kind`
 	Host string `host address`
 	Port int32
 	Status int32
 }

 //single monitor info
 type monitor struct {
 	address string `monitor remote address`
 	conn *grpc.ClientConn `rpc connect`
 	client pb.MonitorServiceClient `service client`
 	ctx context.Context
 	pinged bool
 }

 //client info
 type MonitorClient struct {
 	curNode NodeInfo `current client node info`
 	cbNotify func(NodeInfo) bool `callback for node notify`
 	cbUp func(string, string, int32) `callback for node up`
 	cbDown func(string, string, int32) `callback for node down`
 	monitorMap map[string]*monitor `running monitor map, monitorAddr:monitor`
 	streamMap map[string]*pb.MonitorService_NotifyNodeClient `stream map, monitorAddr:stream`
 	closeChan chan bool `main process close chan`
 	logService *tc.LogService `log service object`
 	sync.Mutex `internal data locker`
 }

 //construct
func NewMonitorClient() *MonitorClient {
	//self init
	this := &MonitorClient{
		monitorMap:make(map[string]*monitor),
		streamMap:make(map[string]*pb.MonitorService_NotifyNodeClient),
		closeChan:make(chan bool),
		logService:nil,
	}

	//set log dir
	this.setLog("logs", "client")

	//run main process
	go this.runMainProcess()
	return this
}

//clean up
func (m *MonitorClient) CleanUp() {
	m.closeChan <- true
	if m.logService != nil {
		m.logService.Close()
	}
	time.Sleep(time.Second/20)
}

//set current client node info
func (m *MonitorClient) SetClient(kind, host string, port int32)  {
	node := NodeInfo{
		Kind:kind,
		Host:host,
		Port:port,
	}
	m.curNode = node
}

//set node notify call back
func (m *MonitorClient) SetNotifyFunc(f func(NodeInfo)bool) {
	m.cbNotify = f
}

//func (m *MonitorClient) SetUpFunc(f func(kind, host string, port int32)) {
//	m.cbUp = f
//}
//
//func (m *MonitorClient) SetDownFunc(f func(kind, host string, port int32)) {
//	m.cbDown = f
//}

//get batch nodes
func (m *MonitorClient) GetBatchNodes(kind string) []NodeInfo {
	var (
		result = make([]NodeInfo, 0)
		pbResult = &pb.NodesResult{}
		err error
	)

	if len(m.monitorMap) <= 0 {
		return result
	}

	query := &pb.NodesQuery{
		Kind:kind,
	}
	for _, monitor := range m.monitorMap {
		if monitor.conn == nil {
			continue
		}
		pbResult, err = monitor.client.QueryNodes(monitor.ctx, query)
		if err != nil {
			log.Println("Query nodes failed, err:", err.Error())
			continue
		}
	}

	if pbResult == nil {
		return result
	}

	//copy result for return
	for _, node := range pbResult.NodeList {
		result = append(result, NodeInfo{
			Kind:node.Kind,
			Host:node.Host,
			Port:node.Port,
			Status:node.Status,
		})
	}
	return result
}

//add monitor
//this host and port is monitor side
func (m *MonitorClient) AddMonitor(host string, port int) bool {
	if host == "" || port <= 0 {
		return false
	}
	address := fmt.Sprintf("%s:%d", host, port)
	//try connect monitor
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Println("Can not connect monitor ", address, ", error:", err.Error())
		return false
	}

	//create service client
	client := pb.NewMonitorServiceClient(conn)

	//init monitor
	monitor := &monitor{
		address:address,
		conn:conn,
		client:client,
		ctx:context.Background(),
		pinged:false,
	}

	//add monitor into map
	m.Lock()
	m.monitorMap[address] = monitor
	m.Unlock()

	//begin ping monitor
	go m.ping(monitor, false)

	return true
}

///////////////
//private func
//////////////

//set log dir
func (m *MonitorClient) setLog(logPath, logPrefix string) {
	if m.logService == nil {
		m.logService = tc.NewLogService(logPath, logPrefix)
	}
}

//ping monitor
//this host and port is client side
func (m *MonitorClient) ping(mr *monitor, isReConn bool) bool {
	var err error
	var stream pb.MonitorService_NotifyNodeClient

	if isReConn {
		if mr.conn != nil {
			m.Lock()
			mr.conn.Close()
			mr.conn = nil
			m.Unlock()
		}

		//try connect monitor
		mr.conn, err = grpc.Dial(mr.address, grpc.WithInsecure())
		if err != nil {
			log.Println("Can not reconnect monitor ", mr.address, ", error:", err.Error())
		}

		//reset client & conn
		mr.client = pb.NewMonitorServiceClient(mr.conn)
		time.Sleep(time.Second/10)
	}

	//create stream of both side
	for {
		stream, err = mr.client.NotifyNode(mr.ctx)
		if err == nil {
			break
		}
		//some error happened
		log.Println("Create stream with monitor ", mr.address, " failed, err:", err.Error())
		time.Sleep(time.Second)
	}

	//send stream data to monitor
	node := &pb.NodeInfo{
		Kind:m.curNode.Kind,
		Host:m.curNode.Host,
		Port:m.curNode.Port,
	}
	err = stream.Send(node)
	if err != nil {
		log.Println("Send stream to monitor ", mr.address, " failed, err:", err.Error())
		return false
	}

	//begin receive stream from monitor
	go m.receiveMonitorStream(stream)

	return true
}

//receive stream from monitor
func (m *MonitorClient) receiveMonitorStream(stream pb.MonitorService_NotifyNodeClient) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Monitor data EOF")
			continue
		}
		if err != nil {
			log.Println("Receive failed, err:", err.Error())
			break
		}

		//call related cb func for diff status
		m.cbNotify(NodeInfo{
			Kind:in.Kind,
			Host:in.Host,
			Port:in.Port,
			Status:in.Status,
		})
		time.Sleep(time.Second/10)
	}
}

//internal data clean up
func (m *MonitorClient) dataCleanUp() bool {
	for address, monitor := range m.monitorMap {
		if monitor.conn != nil {
			log.Println("close ", address, " rpc connect")
			monitor.conn.Close()
		}
		delete(m.monitorMap, address)
	}
	return true
}

//check monitor connect status process
func (m *MonitorClient) checkMonitorStatus() bool {
	var monitors = len(m.monitorMap)
	if monitors <= 0 {
		return false
	}
	for _, monitor := range m.monitorMap {
		state := monitor.conn.GetState().String()
		if state == "TRANSIENT_FAILURE" || state == "SHUTDOWN" {
			m.ping(monitor, true)
		}
	}
	return true
}


//main process
func (m *MonitorClient) runMainProcess() {
	var (
		tick = time.Tick(time.Second * MonitorCheckRate)
		needQuit bool
	)
	for {
		if needQuit {
			break
		}
		select {
		case <- tick:
			//check monitors
			m.checkMonitorStatus()
		case <- m.closeChan:
			needQuit = true
		}
	}

	//clean up
	m.dataCleanUp()
}