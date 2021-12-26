package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"unsafe"
)

type ChatMsg struct {
	From, To, Msg string
}

type ClientMsg struct {
	To      string  `json:"To"`      //接收者
	Msg     string  `json:"nsg"`     //消息
	Datalen uintptr `json:"datalen"` //消息长度
}

var chan_msgcenter chan ChatMsg
var mapName2CliAddr map[string]string //昵称->remoteaddr
var mapCliaddr2Clients map[string]net.Conn

func main() {
	mapCliaddr2Clients = make(map[string]net.Conn)
	mapName2CliAddr = make(map[string]string)
	chan_msgcenter = make(chan ChatMsg)

	//绑定ip端口启动监听
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Panic("failed to listen", err)
	}
	defer listener.Close()

	//启动消息中心
	go msg_center()

	//等待新连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept", err)
			break
		}

		go handleConn(conn)
	}
}

func msg_center() {
	for {
		msg := <-chan_msgcenter
		go send_msg(msg)
	}
}

func handleConn(conn net.Conn) {
	from := conn.RemoteAddr().String()
	mapCliaddr2Clients[from] = conn
	msg := ChatMsg{from, "all", from + "->login"}
	chan_msgcenter <- msg
	defer logout(conn, from)

	//分析消息 通知
	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("failed to read", err, from)
			break
		}
		if n > 0 {
			var climsg ClientMsg
			err = json.Unmarshal(buf[:n], &climsg)
			if err != nil {
				fmt.Println("failed to unmashal", err, string(buf[:n]))
				continue
			}
			if climsg.Datalen != unsafe.Sizeof(climsg) {
				fmt.Println("Msg format error", climsg)
				continue
			}

			//组织一个消息到消息中心
			chatmsg := ChatMsg{from, "all", climsg.Msg}
			switch climsg.To {
			case "all":
			case "set":
				mapName2CliAddr[climsg.Msg] = from
				chatmsg.Msg = from + "set name=" + climsg.Msg + "success"
				chatmsg.From = "server"
			default:
				chatmsg.To = climsg.To
			}

			chan_msgcenter <- chatmsg
		}
	}
}

func logout(conn net.Conn, from string) {
	defer conn.Close()
	delete(mapCliaddr2Clients, from)
	msg := ChatMsg{from, "all", from + "->logout"}
	chan_msgcenter <- msg
}

func send_msg(msg ChatMsg) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Failed to marshal", err, msg)
		return
	}
	if msg.To == "all" {
		//广播
		for _, v := range mapCliaddr2Clients {
			//广播不给自己发
			if msg.From != v.RemoteAddr().String() {
				v.Write(data)
			}
		}
	} else {
		//私信
		//通过昵称查找remoteaddr
		from, ok := mapName2CliAddr[msg.To]
		if !ok {
			fmt.Println("user not exist", msg.To)
			return
		}
		//通过remoteaddr查找conn
		conn, ok := mapCliaddr2Clients[from]
		if !ok {
			fmt.Println("client not exists", from, msg.To)
			return
		}

		conn.Write(data)
	}
}
