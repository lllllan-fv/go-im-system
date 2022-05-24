package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//	在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//	消息广播
	Message chan string
}

func newServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听 Server 广播，一旦有消息就发送给在线用户
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()

	}
}

// 消息广播
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 建立连接
func (this *Server) Handler(conn net.Conn) {

	user := NewUser(conn)

	// 用户上线，将用户加入 OnlineMap 中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播当前用户上线消息
	this.BroadCast(user, "已上线")

	// 接收用户发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 获取用户的消息（去掉'\n'）
			msg := string(buf[:n-1])

			this.BroadCast(user, msg)
		}
	}()

	// 阻塞当前 handler，B友说不写也没关系。没明白是什么道理
	select {}
}

func (this *Server) Start() {
	// 监听端口
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	// 报错退出
	if err != nil {
		fmt.Println("net.Listener err:", err)
		return
	}

	// 结束时关闭监听
	defer listener.Close()

	// 启动监听
	go this.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		go this.Handler(conn)
	}

}
