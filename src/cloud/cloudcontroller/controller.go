package cloudcontroller

import (
	"dockerapigo/src/cloud/cloudhub"
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/node"
	"log"
	"sync"
)

type CloudController struct {
	hub        *cloudhub.Hub
	cmdChannel chan message.Message
	Nodes      sync.Map
	Services   sync.Map
}

func NewCloudController() *CloudController {
	cc := &CloudController{
		hub:        cloudhub.NewHub(),
		cmdChannel: make(chan message.Message, 10),
	}
	return cc
}

func (cc *CloudController) AsyncMessageHandler() {
	for {
		select {
		//Command from user
		case cmd := <-cc.cmdChannel:
			cc.CommandHandler(cmd)
		//Message from nodes
		case msg := <-cc.hub.AsyncMessage:
			log.Println(msg)
		case n := <-cc.hub.RegisterToHub:
			cc.Nodes.Store(n.NodeID, n)
			log.Println(n)
		}
	}
}

func (cc *CloudController) CommandHandler(cmd message.Message) {
	resType := cmd.GetSource()
	resOp := cmd.GetOperation()
	switch resType {
	case message.ResourceTypePod:
		switch resOp {
		case message.InsertOperation:
			cc.StartPod(message.ReadPodConfigMap(&cmd))
		}
		break
	case message.ResourceTypeNode:
		break
	case message.ResourceTypeService:
		switch resOp {
		case message.InsertOperation:
			cc.NewService(message.ReadServiceConfigMap(&cmd))
		}
		break
	case message.ResourceTypePodlist:
		break
	}
}

func (cc *CloudController) PodStatusQuiry(groupID string, nodeID string, podID string) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, groupID, nodeID, message.ResourceTypePod, message.QueryOperation)
	msg.FillBody(message.PodConfig{PodName: podID})
	msg.SetSync()
	reply := cc.hub.SendMessageSync(*msg)
	if reply.GetOperation() == message.ResponseErrorOperation {
		log.Println("Error Occur")
		return
	}
	log.Println(message.ReadPodQuiryResponse(&reply))
}

func (cc *CloudController) StartPod(cfg message.PodConfig) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, cfg.Group, cfg.Node, message.ResourceTypePod, message.InsertOperation)
	cfg.HostsCfg = cc.GetServiceString()
	log.Println("Host config:", cfg.HostsCfg)
	msg.SetSync()
	msg.FillBody(cfg)
	reply := cc.hub.SendMessageSync(*msg)
	log.Println(message.ReadPodCreateResponse(&reply))
}

func (cc *CloudController) NodeStatusQuiry() {

}

func (cc *CloudController) NewService(cfg message.Service) {
	if no, ok := cc.Nodes.Load(cfg.Node); ok {
		ip := no.(node.Node).LocalIP
		cfg.LocalIP = ip
	}
	cc.Services.Store(cfg.Name, cfg)
	log.Println("New Service:", cfg)
}

func (cc *CloudController) GetServiceString() []string {
	var res []string
	cc.Services.Range(func(key, value interface{}) bool {
		ret := ""
		s := value.(message.Service)
		ret = ret + s.Name + ":" + s.LocalIP
		res = append(res, ret)
		return true
	})
	return res
}

func (cc *CloudController) RunController() {
	go cc.hub.RunHub()
	cc.AsyncMessageHandler()
}
