package service

import (
	"errors"
	"github.com/andyzhou/monitor/base"
	"github.com/andyzhou/monitor/face"
	pb "github.com/andyzhou/monitor/pb"
	"golang.org/x/net/context"
	"io"
	"log"
	"time"
)

/*
 * rpc node service call back
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //rpc node info
type RpcNode struct {
	base.Basic
}

//node notify monitor
//this called from outside
func (r *RpcNode) NotifyNode(stream pb.MonitorService_NotifyNodeServer) error {
	var in *pb.NodeInfo
	var err error
	var tips string

	//get tag by stream
	tag, ok := r.GetConnTagFromContext(stream.Context())
	if !ok {
		tips = "Can't get tag from node stream."
		log.Println(tips)
		return errors.New(tips)
	}

	//get active node face
	activeNodeFace := face.RunInterFace.GetActiveNode()
	if activeNodeFace == nil {
		tips = "Can't get active node face."
		log.Println(tips)
		return errors.New(tips)
	}

	//try receive stream data from node
	for {
		in, err = stream.Recv()
		if err == io.EOF {
			log.Println("Read done")
			return nil
		}
		if err != nil {
			log.Println("Read error:", err.Error())
			return err
		}

		//sync node into running map
		activeNodeFace.NodeNotify(tag.RemoteAddr.String(), in, stream)

		time.Sleep(time.Second/20)
	}
	return nil
}

//get batch nodes, support filter by kind
func (r *RpcNode) QueryNodes(ctx context.Context, in *pb.NodesQuery) (*pb.NodesResult, error)  {
	//get related nodes
	activeNodeFace := face.RunInterFace.GetActiveNode()
	if activeNodeFace == nil {
		return nil, errors.New("get active node face failed")
	}
	nodeMap := activeNodeFace.GetNodes(in.GetKind())

	//format result
	result := &pb.NodesResult{
		NodeList:make([]*pb.NodeInfo, 0),
	}
	for _, node := range nodeMap {
		result.NodeList = append(result.NodeList, &pb.NodeInfo{
			Kind:node.Kind,
			Host:node.Host,
			Port:node.Port,
			Status:face.NodeStatUp,
		})
	}
	return result, nil
}

