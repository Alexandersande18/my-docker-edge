package cloudcontroller

import (
	"dockerapigo/src/cloud/cloudhub"
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/types"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
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
				cc.NodeStatusQuery(n.NodeID)
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
	pod := message.ReadPodQueryResponse(&reply)
	fmt.Printf("-------Pod Summary-------\n")
	fmt.Println("Node   :", nodeID)
	fmt.Println("PodName:", pod.PodName)
	fmt.Println("ID     :", pod.ID)
	fmt.Println("Status :", pod.Status)
	fmt.Println("Image  :", pod.Image)
	fmt.Println("PortMap:", pod.PortMap)
	//log.Println(message.ReadPodQueryResponse(&reply))
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

func (cc *CloudController) PrintPodStatus(nodeID string, podID string) {
	//if no, nodeExist := cc.Nodes.Load(nodeID); nodeExist {
	//	node := no.(*types.Node)
	//	if pod, podExist := node.Pods[podID]; podExist {
	//
	//	}
	//
	//} else {
	//	log.Println("Node", nodeID, "does not exist!")
	//}
}

func (cc *CloudController) StartPod(cfg message.PodConfig) {
	if no, nodeExist := cc.Nodes.Load(cfg.Node); nodeExist {
		node := no.(*types.Node)
		if _, podExist := node.Pods[cfg.PodName]; podExist {
			log.Println("Pod", cfg.PodName, "Already exist")
			log.Println("Updating Pod", cfg.PodName)
			time.Sleep(5 * time.Second)
			log.Println("Update Success")
			return
		}
		msg := message.NewMessage(config.MasterID)
		msg.BuildRouter(config.MasterID, cfg.Group, cfg.Node, message.ResourceTypePod, message.InsertOperation)
		cfg.HostsCfg = cc.makeServiceString(cfg.EnvCfg)
		log.Println("Host config:", cfg.HostsCfg)
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

func (cc *CloudController) PodUpdate(cfg message.PodConfig) {
	if no, nodeExist := cc.Nodes.Load(cfg.Node); nodeExist {
		node := no.(*types.Node)
		if _, podExist := node.Pods[cfg.PodName]; podExist {

		} else {
			log.Println("Pod", cfg.PodName, "does not exist!")
		}
	} else {
		log.Println("Node", cfg.Node, "does not exist!")
	}
}

func (cc *CloudController) PodRemove(groupID, nodeID, podID string) {
	if no, nodeExist := cc.Nodes.Load(nodeID); nodeExist {
		node := no.(*types.Node)
		delete(node.Pods, podID)
		msg := message.NewMessage(config.MasterID)
		msg.SetRoute(config.MasterID, groupID, nodeID).SetResourceOperation(message.ResourceTypePod, message.DeleteOperation)
		msg.FillBody(podID)
		reply := cc.hub.SendMessageSync(*msg)
		a := reply.GetContentRaw().(string)
		log.Println("Pod remove reply", a)
		log.Println("Pod remove success")
	} else {
		log.Println("Node", nodeID, "does not exist!")
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
			service, _ := cc.Services.Load(strings.Split(env.Value, ":")[0])
			serviceString = append(serviceString, strings.Split(env.Value, ":")[0]+":"+service.(message.Service).LocalIP)
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
