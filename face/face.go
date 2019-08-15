package face

/**
 * Inter face
 */

 //face info
 type InterFace struct {
 	rpcStat *RpcStat
 	activeNode *ActiveNode
 }

 //declare global variable
 var RunInterFace *InterFace

 //construct
func NewInterFace() *InterFace {
	this := &InterFace{
		rpcStat:NewRpcStat(),
		activeNode:NewActiveNode(),
	}
	return this
}

//quit
func (f *InterFace) Quit() {
	f.activeNode.Quit()
}

//get relate face
func (f *InterFace) GetRpcStat() *RpcStat {
	return f.rpcStat
}

func (f *InterFace) GetActiveNode() *ActiveNode {
	return f.activeNode
}
