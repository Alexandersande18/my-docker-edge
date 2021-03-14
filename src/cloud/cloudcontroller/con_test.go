package cloudcontroller

import (
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/node"
	"log"
	"testing"
)

func TestController(t *testing.T) {
	cc := NewCloudController()
	n := node.Node{
		NodeID:  "c01",
		LocalIP: "192.168.116.12",
		Group:   "0",
		Status:  "",
	}
	cc.Nodes.Store(n.NodeID, n)
	cc.NewService(message.Service{
		Name:     "mongo",
		Node:     "c01",
		Pod:      "mongodb",
		Protocol: "tcp",
		Port:     "17017",
	})

	log.Println(cc.GetServiceString())
	//cc.RunController()
}
