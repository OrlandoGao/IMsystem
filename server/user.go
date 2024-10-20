package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn

	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

//上线业务
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播上线消息
	this.server.BroadCast(this, "已上线")
}

//下线业务
func (this *User) Offline() {
	//用户下线,从Map删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播下线
	this.server.BroadCast(this, "已下线")
}

//给User单独发送消息
func (this *User) SendMsg (msg string) {
	this.conn.Write([]byte(msg))
}

//处理消息业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些
		
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
 		}
		this.server.mapLock.Unlock()
	
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|newname
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("该用户名已被使用")
		}else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("更名成功为: " + newName + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|"{
		//消息格式： to|userName|msg

		//1 获取对方用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不对 请使用 : \" to|remoteName|...\" 格。式\n")
			return
		}

		//2 根据用户名 得到目标的User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户不存在或不在线\n")
			return
		}

		//3 获取消息
		content := strings.Split(msg,"|")[2]
		if content == "" {
			this.SendMsg("无消息内容请重发")
			return
		}
		remoteUser.SendMsg("\n以下为私聊:\n")
		remoteUser.SendMsg(this.Name + "对您说: " + content + "\n\n")

	}else {
	this.server.BroadCast(this, msg)
	}
}

//监听当前User channel的方法，一旦有消息就发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}