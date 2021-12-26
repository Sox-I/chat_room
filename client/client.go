package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"unsafe"
)

//与客户端通信结构
type ClientMsg struct {
	To      string  `json:"To"`      //接收者
	Msg     string  `json:"nsg"`     //消息
	Datalen uintptr `json:"datalen"` //消息长度
}

func Help() {
	fmt.Println("1. set:your name")
	fmt.Println("2. all: send your message to all users")
	fmt.Println("3. anyone:your msg -- private message")
}

func handle_conn(conn net.Conn) {
	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Panic("failed to read", err)
		}
		fmt.Println(string(buf[:n]))
		fmt.Printf("Bill's chat>")
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		log.Panic("failed to connect", err)
	}
	defer conn.Close()

	go handle_conn(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to my chatroom\n")
	Help()
	for {
		fmt.Printf("Bill's chat>")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Panic("failed to read", err)
		}
		msg = strings.Trim(msg, "\r\n") //取消换行回车

		//quit
		if msg == "quit" {
			fmt.Println("bye bye")
			break
		}
		if msg == "help" {
			Help()
			continue
		}

		//消息处理
		msgs := strings.Split(msg, ":")
		if len(msgs) == 2 {
			var climsg ClientMsg
			climsg.To = msgs[0]
			climsg.Msg = msgs[1]
			climsg.Datalen = unsafe.Sizeof(climsg)

			//转换json为[]byte
			data, err := json.Marshal(climsg)
			if err != nil {
				fmt.Println("failed to marshal", err, climsg)
				continue
			}
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println("failed to write", err, climsg)
				break
			}
		}
	}
}
