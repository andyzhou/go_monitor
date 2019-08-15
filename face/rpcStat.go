package face

import (
	"google.golang.org/grpc/stats"
	"sync"
)

/*
 * rcp stat data face
 */

 //stat info
 type RpcStat struct {
 	connMap map[*stats.ConnTagInfo]string
 	sync.RWMutex
 }

 //construct
func NewRpcStat() *RpcStat {
	//self init
	this := &RpcStat{
		connMap:make(map[*stats.ConnTagInfo]string),
	}
	return this
}

////////
//api
///////

//remove connect
func (f *RpcStat) RemoveConn(tag *stats.ConnTagInfo) bool {
	if tag == nil {
		return false
	}
	//remove with locker
	f.Lock()
	defer f.Unlock()
	delete(f.connMap, tag)
	return true
}

//add connect
func (f *RpcStat) AddConn(tag *stats.ConnTagInfo, address string) bool {
	if tag == nil || address == "" {
		return false
	}

	//add into map with locker
	f.Lock()
	defer f.Unlock()
	f.connMap[tag] = address

	return true
}