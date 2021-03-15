package cloudcontroller

import (
	"dockerapigo/src/cloud/cloudhub"
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/types"
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
			if msg.GetResource() == message.ResourceTypePod && msg.GetOperation() == message.ResponseOperation {
				log.Println("Pod Query Reply")
				reply := message.ReadPodQuiryResponse(&msg)
				no, _ := cc.Nodes.Load(msg.GetSource())
				node := no.(*types.Node)
				log.Println(*node)
				if pod, ok := node.Pods[reply.PodName]; ok {
					pod.Info = reply
					log.Println("Pod Info refresh", pod.ToString())
				}
			}
			log.Println("Async Message:", msg)
		case n := <-cc.hub.RegisterToHub:
			if old, loaded := cc.Nodes.LoadOrStore(n.NodeID, &n); loaded {
				// Node Registered before, so set Status
				node := old.(*types.Node)
				node.Status = n.Status
				log.Println("Node", *node, "Status Changed")
			} else {
				log.Println("New Node", n)
			}
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
		log.Println("Error Occur", reply.GetContent())
		return
	}
	log.Println(message.ReadPodQuiryResponse(&reply))
}

func (cc *CloudController) AsyncPodStatusQuiry(groupID string, nodeID string, podID string) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, groupID, nodeID, message.ResourceTypePod, message.QueryOperation)
	msg.FillBody(message.PodConfig{PodName: podID})
	cc.hub.SendMessage(*msg)
}

func (cc *CloudController) StartPod(cfg message.PodConfig) {
	if no, nodeExist := cc.Nodes.Load(cfg.Node); nodeExist {
		node := no.(*types.Node)
		if _, podExist := node.Pods[cfg.PodName]; podExist {
			log.Println("Pod", cfg.PodName, "Already exist")
			return
		}
		msg := message.NewMessage(config.MasterID)
		msg.BuildRouter(config.MasterID, cfg.Group, cfg.Node, message.ResourceTypePod, message.InsertOperation)
		cfg.HostsCfg = cc.GetServiceString()
		log.Println("Host config:", cfg.HostsCfg)
		msg.SetSync()
		msg.FillBody(cfg)
		reply := cc.hub.SendMessageSync(*msg)
		response := message.ReadPodCreateResponse(&reply)
		log.Println(response)
		node.Pods[cfg.PodName] = &types.Pod{
			PodName: cfg.PodName,
			NodeID:  cfg.Node,
			Info:    message.PodQuiryResponse{},
		}
	}
}

func (cc *CloudController) NodeStatusQuiry(nodeID string) {
	if nd, ok := cc.Nodes.Load(nodeID); ok {
		log.Println(nd.(*types.Node).ToString())
	} else {
		log.Println(nodeID + "Does not exist")
	}
}

func (cc *CloudController) NewService(cfg message.Service) {
	if no, ok := cc.Nodes.Load(cfg.Node); ok {
		ip := no.(*types.Node).LocalIP
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
