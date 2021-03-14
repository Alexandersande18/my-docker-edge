package edgehub

import (
	"dockerapigo/src/common/message"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

type WsClient struct {
	conn            *websocket.Conn
	addr            *string
	path            string
	sendMsgChan     chan message.Message
	RecvMsgChan     chan message.Message
	RegisterMessage message.Message
	isAlive         bool
	timeout         int
	id              string
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

// 构造函数
func NewWsClientManager(addrIp, addrPort, path string, timeout int, id string) *WsClient {
	addrString := addrIp + ":" + addrPort
	var sendChan = make(chan message.Message, 10)
	var recvChan = make(chan message.Message, 10)
	var conn *websocket.Conn
	NodeIp, _ := GetOutBoundIP()
	return &WsClient{
		addr:            &addrString,
		path:            path,
		conn:            conn,
		sendMsgChan:     sendChan,
		RecvMsgChan:     recvChan,
		isAlive:         false,
		timeout:         timeout,
		id:              id,
		RegisterMessage: *message.NewRegisterMessage(id, NodeIp),
	}
}

// 链接服务端
func (wsc *WsClient) dail() {
	var err error
	u := url.URL{Scheme: "ws", Host: *wsc.addr, Path: wsc.path}
	log.Printf("connecting to %s", u.String())
	wsc.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsc.isAlive = true
	wsc.SendMsg(wsc.RegisterMessage)
	log.Printf("Connected to %s ", u.String())
}

// 发送消息
func (wsc *WsClient) sendMsgThread() {
	go func() {
		for {
			if wsc.conn != nil {
				msg := <-wsc.sendMsgChan
				//err := wsc.conn.WriteMessage(websocket.TextMessage, []byte(msg))
				err := wsc.conn.WriteJSON(msg)
				if err != nil {
					log.Println("write:", err)
					wsc.isAlive = false
					break
				}
			} else {
				break
			}
		}
	}()
}

// 读取消息
func (wsc *WsClient) readMsgThread() {
	go func() {
		for {
			if wsc.conn != nil {
				//_, msg, err := wsc.conn.ReadMessage()
				msg := &message.Message{}
				err := wsc.conn.ReadJSON(msg)
				if err != nil {
					log.Println("read:", err)
					wsc.isAlive = false
					// 出现错误，退出读取，尝试重连
					break
				}
				wsc.RecvMsgChan <- *msg
				log.Println("recv: ", *msg)
			} else {
				break
			}
		}
	}()
}

// 开启服务并重连
func (wsc *WsClient) Start() {
	for {
		if wsc.isAlive == false {
			wsc.dail()
			wsc.sendMsgThread()
			wsc.readMsgThread()
			log.Println("Timeout, trying to recon")
		}
		time.Sleep(time.Duration(wsc.timeout) * time.Second)
	}
}

func (wsc *WsClient) SendMsg(msg message.Message) {
	wsc.sendMsgChan <- msg
}
