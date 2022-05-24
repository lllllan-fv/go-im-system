package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// 用户上线，信息存入 server.OnlineMap
func (this *User) Online() {

	// 用户上线，将用户加入 OnlineMap 中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线，信息从 server.OnlineMap 移出
func (this *User) Offline() {

	// 用户下线，将用户从 OnlineMap 中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播下线消息
	this.server.BroadCast(this, "下线")
}

// 当前用户下显示内容
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineUser := "[" + user.Addr + "]" + user.Name + ":" + "在线"
			this.SendMsg(onlineUser)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.BroadCast(this, msg)
	}
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.SendMsg(msg)
	}
}
