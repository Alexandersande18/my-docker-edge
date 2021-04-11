package cloudcontroller

import (
	"dockerapigo/src/common/types"
	"sync"
)

func getNextNode(nodes *sync.Map) string {
	minPodNum := 1000
	var targetNodeId string
	nodes.Range(func(nodeId, n interface{}) bool {
		node := n.(types.Node)
		if node.Status == types.NodeStatusAlive {
			t := len(node.Pods)
			if t < minPodNum {
				minPodNum = t
				targetNodeId = nodeId.(string)
			}
		}
		return true
	})
	return targetNodeId
}
