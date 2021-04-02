package cloudhub

import (
	"dockerapigo/src/common/config"
	"dockerapigo/src/common/message"
	"dockerapigo/src/common/types"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	clientToId map[*Client]string

	idToClient map[string]*Client

	// Inbound messages from the clients.
	messageToSend   chan message.Message
	messageReceived chan message.Message
	AsyncMessage    chan message.Message
	syncMessage     sync.Map
	// Register requests from the clients.
	register      chan *Client
	RegisterToHub chan types.Node
	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		messageToSend:   make(chan message.Message, 10),
		messageReceived: make(chan message.Message, 100),
		AsyncMessage:    make(chan message.Message, 20),
		register:        make(chan *Client, 10),
		RegisterToHub:   make(chan types.Node, 10),
		unregister:      make(chan *Client, 10),
		clientToId:      make(map[*Client]string),
		idToClient:      make(map[string]*Client),
	}
}

func (h *Hub) RunHub() {
	go h.MessageHandler()
	//go hub.SendMsg()
	var addr = flag.String("addr", ":"+config.MasterPort, "http service address")
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (h *Hub) MessageHandler() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case msg := <-h.messageToSend:
			if _, ok := h.idToClient[msg.GetTarget()]; ok {
				client := h.idToClient[msg.GetTarget()]
				select {
				case client.send <- msg:
				default:
					log.Println("Client.send not ready")
				}
			} else {
				log.Println("no client has id", msg.GetTarget())
				h.syncMsgArrival(msg.GetID(), message.NewClientOfflineErrMessage(msg.GetID()))
			}
		case msg := <-h.messageReceived:
			log.Println("CloudHub Received: ", msg.GetContentRaw())
			if msg.IsSync() {
				h.syncMsgArrival(msg.GetParentID(), &msg)
			} else {
				h.AsyncMessage <- msg
			}
		}
	}
}

func (h *Hub) syncMsgArrival(parentID string, msg *message.Message) {
	if value, ok := h.syncMessage.Load(parentID); ok {
		if value == nil {
			h.syncMessage.Store(parentID, msg)
		}
	}
}

func (h *Hub) syncMsgPolling(ID string) message.Message {
	for {
		if reply, ok := h.syncMessage.Load(ID); ok {
			if reply != nil {
				h.syncMessage.Delete(ID)
				return *(reply).(*message.Message)
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (h *Hub) registerClient(cli *Client) {
	if cli.id == "-1" {
		log.Println("Client has no id")
	}
	h.clientToId[cli] = cli.id
	h.idToClient[cli.id] = cli
	h.RegisterToHub <- types.Node{
		NodeID:  cli.id,
		LocalIP: cli.LocalIP,
		Group:   cli.group,
		Status:  types.NodeStatusAlive,
		Pods:    make(map[string]*types.Pod),
	}
	log.Println("New Client ", cli.id, cli.LocalIP)
}

func (h *Hub) unregisterClient(cli *Client) {
	if _, ok := h.clientToId[cli]; ok {
		log.Println("Client leaving:", *cli)
		id := h.clientToId[cli]
		h.RegisterToHub <- types.Node{
			NodeID:  cli.id,
			LocalIP: cli.LocalIP,
			Group:   cli.group,
			Status:  types.NodeStatusDead,
		}
		delete(h.clientToId, cli)
		delete(h.idToClient, id)
		close(cli.send)
	}
}

func (h *Hub) ClientRegistered(cid string) bool {
	if _, ok := h.idToClient[cid]; ok {
		return true
	} else {
		return false
	}
}

func (h *Hub) SendMessage(msg message.Message) {
	h.messageToSend <- msg
}

func (h *Hub) SendMessageSync(msg message.Message) message.Message {
	ID := msg.GetID()
	msg.SetSync()
	h.syncMessage.Store(ID, nil)
	h.messageToSend <- msg
	return h.syncMsgPolling(ID)
}
