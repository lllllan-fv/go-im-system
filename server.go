package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

	user := NewUser(conn, this)

	user.Online()

	isLive := make(chan bool)

	// 接收用户发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 获取用户的消息（去掉'\n'）
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
			//	当前用户活跃，重置定时器

		case <-time.After(300 * time.Second):
			//	已经超时，强制踢出当前客户端
			user.SendMsg("你被踢了")

			// 释放资源
			close(user.C)

			// 关闭连接
			conn.Close()

			return
		}
	}
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
