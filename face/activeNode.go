package face

import (
	"github.com/andyzhou/monitor/define"
	pb "github.com/andyzhou/monitor/pb"
	"log"
	"sync"
	"time"
)

/*
 * Active node data face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //node status
 const (
 	NodeStatDown = iota
 	NodeStatUp
 )

 //single node info
 type NodeInfo struct {
 	RemoteAddr string
 	Kind string
 	Host string
 	Port int32
 	Status int32
 	Stream *pb.MonitorService_NotifyNodeServer
 }

 //active node info
 type ActiveNode struct {
 	nodeMap map[string]*NodeInfo `running active client node map, remoteAddr:NodeInfo`
 	nodeChan chan NodeInfo `node receiver chan`
 	removeChan chan string `node remove chan`
 	closeChan chan bool
 	sync.RWMutex `internal data locker`
 }

 //construct
func NewActiveNode() *ActiveNode {
	//self init
	this := &ActiveNode{
		nodeMap:make(map[string]*NodeInfo),
		nodeChan:make(chan NodeInfo, define.NodeChanSize),
		removeChan:make(chan string, define.NodeChanSize),
		closeChan:make(chan bool),
	}
	//run main process
	go this.runMainProcess()
	return this
}

//////
//API
//////

//quit
func (a *ActiveNode) Quit() {
	a.closeChan <- true
	time.Sleep(time.Second/20)
}

//check node is exists or not
func (a *ActiveNode) NodeIsExists(remoteAddr string) bool {
	if remoteAddr == "" {
		return false
	}
	_, isOk := a.nodeMap[remoteAddr]
	return isOk
}

//get batch nodes by kind
func (a *ActiveNode) GetNodes(kind string) map[string]*NodeInfo{
	result := make(map[string]*NodeInfo)
	for k, node := range a.nodeMap {
		if kind != "" && kind != node.Kind {
			continue
		}
		result[k] = node
	}
	return result
}

//node info notify
func (a *ActiveNode) NodeNotify(remoteAddr string, inInfo *pb.NodeInfo, stream pb.MonitorService_NotifyNodeServer) (bRet bool) {
	//basic check
	bRet = false
	if remoteAddr == "" || inInfo == nil {
		return
	}

	//try catch panic
	defer func(bRet bool) {
		if err := recover(); err != nil {
			log.Println("ActiveNode::NodeNotify panic happend, err:", err)
			bRet = false
		}
	}(bRet)

	//init new node info
	nodeInfo := NodeInfo{
		RemoteAddr:remoteAddr,
		Kind:inInfo.Kind,
		Host:inInfo.Host,
		Port:inInfo.Port,
		Status:inInfo.Status,
		Stream:&stream,
	}

	//cast to chan
	a.nodeChan <- nodeInfo
	bRet = true
	return
}

//new node up
//func (a *ActiveNode) NodeUp(remoteAddr, kind, host string, port int32, stream pb.MonitorService_NotifyNodeServer) bool {
//	if host == "" || port <= 0 {
//		return false
//	}
//	nodeInfo := NodeInfo{
//		RemoteAddr:remoteAddr,
//		Kind:kind,
//		Host:host,
//		Port:port,
//		Stream:&stream,
//	}
//	//cast to chan
//	a.nodeChan <- nodeInfo
//	return true
//}

//running node down
//need notify other nodes?
func (a *ActiveNode) NodeDown(remoteAddress string) bool {
	if remoteAddress == "" {
		return false
	}
	//cast to chan
	a.removeChan <- remoteAddress
	return true
}

///////////////
//private func
///////////////

//clear all data
func (a *ActiveNode) clearNodes() bool {
	if len(a.nodeMap) <= 0 {
		return false
	}
	a.Lock()
	defer a.Unlock()
	for k, _ := range a.nodeMap {
		delete(a.nodeMap, k)
	}
	return true
}

//notify other nodes
//call this when node up/down
func (a *ActiveNode) notifyOthers(node *NodeInfo, status int32) {
	var err error
	for k, v := range a.nodeMap {
		if k == node.RemoteAddr {
			continue
		}
		//cast to this node
		err = (*v.Stream).Send(&pb.NodeInfo{
			Kind:node.Kind,
			Host:node.Host,
			Port:node.Port,
			Status:status,
		})
		log.Printf("noitify others %v, err:%v\n", k, err)
	}
}

//remove node
func (a *ActiveNode) removeNode(address string) bool {
	if address == "" {
		return false
	}
	a.Lock()
	defer a.Unlock()
	if node, ok := a.nodeMap[address]; ok {
		//remove node
		delete(a.nodeMap, address)
		//notify
		a.notifyOthers(node, NodeStatDown)
	}
	log.Printf("current nodes:%v\n", len(a.nodeMap))
	return true
}

//sync node
func (a *ActiveNode) syncNode(node *NodeInfo) {
	//check
	if node == nil {
		return
	}
	address := node.RemoteAddr

	//notify
	a.notifyOthers(node, node.Status)

	//sync node into running map
	a.Lock()
	defer a.Unlock()
	a.nodeMap[address] = node
	log.Printf("current nodes:%v\n", len(a.nodeMap))
}

//internal main process
func (a *ActiveNode) runMainProcess() {
	var (
		node NodeInfo
		address string
		isOk bool
	)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("ActiveNode:runMainProcess panic, err:%v\n", err)
		}
		//clean up
		a.clearNodes()
		close(a.nodeChan)
		close(a.removeChan)
	}()

	//loop
	for {
		select {
		case node, isOk = <- a.nodeChan://sync node
			if isOk && &node != nil {
				a.syncNode(&node)
			}
		case address, isOk = <- a.removeChan://remove node
			if isOk && address != "" {
				a.removeNode(address)
			}
		case <- a.closeChan:
			return
		}
	}
}

