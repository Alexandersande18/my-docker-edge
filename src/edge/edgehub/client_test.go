package edgehub

import (
	"dockerapigo/src/common/message"
	"sync"
	"testing"
	"time"
)

var clients = make(chan *WsClient, 3)

func RunClient(hostname string) {
	wsc := NewWsClientManager("localhost", "11451", "/ws", 10, hostname)
	clients <- wsc
	wsc.Start()
	var w1 sync.WaitGroup
	w1.Add(1)
	w1.Wait()
}

func dispachMessage() {
	for {
		select {
		case c := <-clients:
			msg := message.NewMessage(c.id)
			msg.SetRoute(c.id, "0", "Master")
			//log.Println("Making: ", msg)
			c.sendMsgChan <- *msg
			clients <- c
		}
		time.Sleep(1 * time.Second)
	}
}

func TestClient(t *testing.T) {
	go RunClient("0")
	go RunClient("1")
	go RunClient("2")
	go dispachMessage()
	time.Sleep(100 * time.Second)
}
