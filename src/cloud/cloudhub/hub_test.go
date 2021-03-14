package cloudhub

import (
	"dockerapigo/src/common/message"
	"flag"
	"log"
	"net/http"
	"testing"
	"time"
)

func Run() {
	flag.Parse()
	hub := NewHub()
	go hub.MessageHandler()
	//go hub.SendMsg()
	var addr = flag.String("addr", ":11451", "http service address")
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func TestHub(t *testing.T) {
	flag.Parse()
	hub := NewHub()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			msg := message.NewMessage("Master")
			msg.BuildRouter("Master", "0", "c01", message.ResourceTypePod, message.QueryOperation)
			msg.FillBody(message.PodConfig{PodName: "t01"})
			log.Println(hub.SendMessageSync(*msg))
		}
	}()
	go hub.MessageHandler()

	var addr = flag.String("addr", ":1145", "http service address")
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
