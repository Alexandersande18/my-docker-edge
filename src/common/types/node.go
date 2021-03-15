package types

import (
	"dockerapigo/src/common/message"
	"fmt"
)

const (
	NodeStatusAlive = "alive"
	NodeStatusDead  = "dead"
)

type Pod struct {
	PodName string
	NodeID  string
	Info    message.PodQuiryResponse
}

type Node struct {
	NodeID  string
	LocalIP string
	Group   string
	Status  string
	Pods    map[string]*Pod
}

func (p *Pod) ToString() string {
	return fmt.Sprint(*p)
}

func (n *Node) ToString() string {
	podstr := ""
	for _, v := range n.Pods {
		podstr += v.ToString()
	}
	return "ID:" + n.NodeID + "\n" + "Local IP:" + n.LocalIP + "\n" + "Status:" + n.Status + "\n" + "Pods:" + podstr
}
