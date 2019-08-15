package service

import (
	"log"
	"google.golang.org/grpc/stats"
	"golang.org/x/net/context"
	"monitor/base"
	"monitor/face"
)

/*
 * RPC stat handler
 * Need apply `TagConn`, `TagRPC`, `HandleConn`, `HandleRPC` methods.
 */

type RpcStat struct {
	basic *base.Basic
}

//construct
func NewRpcStat() *RpcStat {
	this := &RpcStat{
		basic:new(base.Basic),
	}
	return this
}

//connect ctx key info
//type connCtxKey struct{}

//declare global variables
//var connMutex sync.Mutex
var connMap = make(map[*stats.ConnTagInfo]string)

//func getConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
//	tag, ok := ctx.Value(base.ConnCtxKey{}).(*stats.ConnTagInfo)
//	return tag, ok
//}

func (h *RpcStat) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	log.Println("TagConn, from address:", info.RemoteAddr.String())
	return context.WithValue(ctx, base.ConnCtxKey{}, info)
}

func (h *RpcStat) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	log.Println("TagRPC, method name:", info.FullMethodName)
	return ctx
}

func (h *RpcStat) HandleConn(ctx context.Context, s stats.ConnStats) {
	//get connect tag from context
	tag, ok := h.basic.GetConnTagFromContext(ctx)
	if !ok {
		log.Fatal("can not get conn tag")
		return
	}

	//get rpc stat face
	rpcStatFace := face.RunInterFace.GetRpcStat()
	if rpcStatFace == nil {
		return
	}

	//get active node face
	activeNodeFace := face.RunInterFace.GetActiveNode()
	if activeNodeFace == nil {
		return
	}

	//do relate opt by connect stat type
	switch s.(type) {
	case *stats.ConnBegin:
		rpcStatFace.AddConn(tag, tag.RemoteAddr.String())
	case *stats.ConnEnd:
		//node down
		activeNodeFace.NodeDown(tag.RemoteAddr.String())
		//remove connect
		rpcStatFace.RemoveConn(tag)
	default:
		log.Printf("illegal ConnStats type\n")
	}
}

func (h *RpcStat) HandleRPC(ctx context.Context, s stats.RPCStats) {
	//fmt.Println("HandleRPC, IsClient:", s.IsClient())
}

