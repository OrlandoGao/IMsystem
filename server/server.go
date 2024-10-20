package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)


type Server struct {
	IP string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个server的端口
func NewServer(ip string, port int) *Server{
	server := Server{
		IP: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return &server
}

//监听Message 广播消息的goroutine，有消息就发送给全部的在线Uesr
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		//发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//用户上线
	user := NewUser(conn, this)
	
	user.Online()

	//监听用户是否活跃的channel
	islive := make(chan bool)

	//新协程
	go func ()  {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil {
				fmt.Println("Conn Read err:", err)
				return
			}
			//提取用户消息（去掉'\n'）
			msg := string(buf[:n - 1])
			//处理消息
			user.DoMessage(msg)

			islive <-true
		}
	}()
	
	//当前handler阻塞
	for {
		select {
		case <-islive:
			//活跃的，重置定时器
			//不做 激活select 更新下面计时器
		case <-time.After(time.Second * 300):
			//已经超时将当前User强制的关闭
			user.SendMsg("被踢啦\n")

			//销毁资源
			close(user.C)

			conn.Close()
			//退出当前的Handler
			return 
		}

	}
}

//启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}

	//close listen socket
	defer listener.Close()
}
