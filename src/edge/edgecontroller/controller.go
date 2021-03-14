package edgecontroller

import (
	"context"
	"dockerapigo/src/api"
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/edge/edgehub"
	"github.com/docker/docker/client"
	"log"
	"sync"
)

type EdgeController struct {
	wsclient  *edgehub.WsClient
	ctx       context.Context
	apiClient *client.Client
	nodeId    string
}

func NewEdgeController(masterIp string, masterPort string, nodeID string, hostApiUrl string, ctx context.Context) *EdgeController {
	wsclient := edgehub.NewWsClientManager(masterIp, masterPort, "/ws", config.ClientTimeout, nodeID)
	apiClient, err := api.NewApiClient(hostApiUrl)
	if err != nil {
		log.Println(err)
	}
	return &EdgeController{
		wsclient:  wsclient,
		ctx:       ctx,
		apiClient: apiClient,
		nodeId:    nodeID,
	}
}

func (ec *EdgeController) Run() {
	go ec.MessageHandler()
	ec.wsclient.Start()
	var w1 sync.WaitGroup
	w1.Add(1)
	w1.Wait()
}

func (ec *EdgeController) podCreate(msg *message.Message) message.Message {
	configMap := message.ReadPodConfigMap(msg)
	cid, err := api.RunCbyName(ec.ctx, ec.apiClient, configMap.ImageName, configMap.PodName, configMap.PortsMap,
		configMap.MountsMap, configMap.EnvCfg, configMap.HostsCfg)
	if err != nil {
		log.Println(err)
	}
	res := message.PodCreateResponse{
		Cid:     cid,
		Success: true,
	}
	return *message.NewRespByMessage(msg, res)
}

func (ec *EdgeController) podStop(msg *message.Message) message.Message {
	configMap := message.ReadPodConfigMap(msg)
	res, err := api.StopCbyName(ec.ctx, ec.apiClient, configMap.PodName)
	if err != nil {
		log.Println(err)
	}
	log.Println(res)
	return *message.NewRespByMessage(msg, res)
}

func (ec *EdgeController) podStatus(msg *message.Message) message.Message {
	configMap := message.ReadPodConfigMap(msg)
	res, err := api.GetContainerStatus(ec.ctx, ec.apiClient, configMap.PodName)
	if err != nil {
		log.Println(err)
	}
	reply := message.PodQuiryResponse{
		Status:  res.State.Status,
		Image:   res.Image,
		PortMap: api.GetPortMapString(res.HostConfig.PortBindings),
	}
	return *message.NewRespByMessage(msg, reply)
}

func (ec *EdgeController) handlePodOp(msg message.Message) {
	resOp := msg.Router.Operation
	switch resOp {
	case message.InsertOperation:
		resp := ec.podCreate(&msg)
		ec.wsclient.SendMsg(resp)
		break
	case message.DeleteOperation:
		resp := ec.podStop(&msg)
		ec.wsclient.SendMsg(resp)
		break
	case message.QueryOperation:
		resp := ec.podStatus(&msg)
		ec.wsclient.SendMsg(resp)
		break
	case message.UpdateOperation:
		break
	}
}

func (ec *EdgeController) handleNodeOp(msg message.Message) {
	resOp := msg.Router.Operation
	switch resOp {
	case message.InsertOperation:
		break
	case message.DeleteOperation:
		break
	case message.UpdateOperation:
		break
	}
}

func (ec *EdgeController) MessageHandler() {
	for {
		msg := <-ec.wsclient.RecvMsgChan
		resType := msg.Router.Resource
		switch resType {
		case message.ResourceTypePod:
			ec.handlePodOp(msg)
			break
		case message.ResourceTypeNode:
			ec.handleNodeOp(msg)
			break
		case message.ResourceTypeNodeStatus:
			break
		case message.ResourceTypePodlist:
			break
		}
	}
}
