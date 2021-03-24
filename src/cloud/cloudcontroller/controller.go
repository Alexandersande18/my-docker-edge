package cloudcontroller

import (
	"dockerapigo/src/cloud/cloudhub"
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/types"
	"fmt"
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
			log.Println("Async Message:", msg)
			if msg.GetResource() == message.ResourceTypePodStatus && msg.GetOperation() == message.ResponseOperation {
				//log.Println("PodStatus Query Reply")
				reply := message.ReadPodQueryResponse(&msg)
				no, _ := cc.Nodes.Load(msg.GetSource())
				node := no.(*types.Node)
				log.Println(*node)
				if pod, ok := node.Pods[reply.PodName]; ok {
					pod.Info = reply
					log.Println("Pod Info refresh,", pod.ToString())
				}
			} else if msg.GetResource() == message.ResourceTypePodlist && msg.GetOperation() == message.ResponseOperation {
				reply := message.ReadPodListResponse(&msg)
				no, _ := cc.Nodes.Load(reply.NodeID)
				log.Println("PodList Query Reply", reply)
				node := no.(*types.Node)
				for _, pod := range reply.Pods {
					if _, ok := node.Pods[pod]; !ok {
						node.Pods[pod] = &types.Pod{
							PodName: pod,
							NodeID:  node.NodeID,
							Info:    message.PodQueryResponse{},
						}
						log.Println("Pods Refresh", node.Pods)
					}
				}
			}

		case n := <-cc.hub.RegisterToHub:
			if old, loaded := cc.Nodes.LoadOrStore(n.NodeID, &n); loaded {
				// Node Registered before, so set Status
				node := old.(*types.Node)
				node.Status = n.Status
				log.Println("Node", *node, "Status Changed")
				if node.Status == types.NodeStatusAlive {
					cc.NodeStatusQuery(node.NodeID)
				}
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

func (cc *CloudController) PodStatusQuery(groupID string, nodeID string, podID string) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, groupID, nodeID, message.ResourceTypePodStatus, message.QueryOperation)
	msg.FillBody(message.PodConfig{PodName: podID})
	msg.SetSync()
	reply := cc.hub.SendMessageSync(*msg)
	if reply.GetOperation() == message.ResponseErrorOperation {
		log.Println("Error Occur", reply.GetContent())
		return
	}
	log.Println(message.ReadPodQueryResponse(&reply))
}

func (cc *CloudController) AsyncPodStatusQuery(groupID string, nodeID string, podID string) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, groupID, nodeID, message.ResourceTypePodStatus, message.QueryOperation)
	msg.FillBody(message.PodConfig{PodName: podID})
	cc.hub.SendMessage(*msg)
}

func (cc *CloudController) AsyncPodListQuery(groupID string, nodeID string) {
	msg := message.NewMessage(config.MasterID)
	msg.BuildRouter(config.MasterID, groupID, nodeID, message.ResourceTypePodlist, message.QueryOperation)
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
		cfg.HostsCfg = cc.makeServiceString(cfg.EnvCfg)
		log.Println("Host config:", cfg.HostsCfg)
		msg.SetSync()
		msg.FillBody(cfg)
		reply := cc.hub.SendMessageSync(*msg)
		response := message.ReadPodCreateResponse(&reply)
		log.Println(response)
		node.Pods[cfg.PodName] = &types.Pod{
			PodName: cfg.PodName,
			NodeID:  cfg.Node,
			Info:    message.PodQueryResponse{},
		}
	} else {
		log.Println("Node", cfg.Node, "does not exist!")
	}
}

func (cc *CloudController) NodeStatusQuery(nodeID string) {
	if nd, ok := cc.Nodes.Load(nodeID); ok {
		cc.AsyncPodListQuery("0", nodeID)
		fmt.Println(nd.(*types.Node).ToString())
	} else {
		log.Println(nodeID + " Does not exist")
	}
}

func (cc *CloudController) NodeList() {
	cc.Nodes.Range(func(key, value interface{}) bool {
		fmt.Printf("-------Nodes-------\n%s\n", value.(*types.Node).ToString())
		fmt.Println()
		return true
	})
}

func (cc *CloudController) makeServiceString(envVars []message.EnvVar) []string {
	var serviceString []string
	for _, env := range envVars {
		if env.Type == "service" {
			service, _ := cc.Services.Load(env.Value)
			serviceString = append(serviceString, env.Value+":"+service.(message.Service).LocalIP)
		}
	}
	return serviceString
}

func (cc *CloudController) NewService(cfg message.Service) {
	if no, ok := cc.Nodes.Load(cfg.Node); ok {
		ip := no.(*types.Node).LocalIP
		cfg.LocalIP = ip
	}
	cc.Services.Store(cfg.Name, cfg)
	log.Println("New Service:", cfg)
}

func (cc *CloudController) RunController() {
	go cc.hub.RunHub()
	cc.AsyncMessageHandler()
}
