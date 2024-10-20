package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP string
	ServerPort int
	Name string
	conn net.Conn
	flag int //当前client的模式
}

var serverIP string
var serverPort int

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIP: serverIp,
		ServerPort: serverPort,
		flag: 999,
	}


	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}

	client.conn = conn

	//返回对象
	return client
}

func (this *Client) Menu() bool {
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")
	var flag int
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法的选项<<<<<<<")
		return false
	}
}

//把服务器回复消息打印出来
func (client *Client) DealResponse() {
	//一旦client.conn 有输出就直接拷贝到标准输出上 永久阻塞
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Update() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.write error: ", err)
		return false
	}

	return true
}

func (client *Client) PublichChat() {
	var chatMsg string
	fmt.Println(">>>>>>请输入聊天信息：(exit退出)")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_,err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.write error: ",err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>>请输入聊天信息：(exit退出)")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) SelectUesrs() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write error: ", err)
		return
	}
}

//私聊模式
func (client *Client) PrivateChat() {
	var remotUser string
	var Msg string
	client.SelectUesrs()
	fmt.Println(">>>>>请输入用户名 (exit退出)")
	fmt.Scanln(&remotUser)
	
	for remotUser != "exit" {
		fmt.Println(">>>>>请输入消息 (exit退出)")
		fmt.Scanln(&Msg)

		for Msg != "exit" {
			
			if len(Msg) != 0 {
			sendMsg := "to|" + remotUser + "|" + Msg + "\n\n"
			_,err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.write error: ",err)
				break
			}
		}
		
		Msg = ""
		fmt.Println(">>>>>请输入消息 (exit退出)")
		fmt.Scanln(&Msg)		
		}
	}

	client.SelectUesrs()
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.Menu() != true {}

		//根据不同的模式判断
		switch client.flag {
		case 1:
			client.PublichChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.Update()
			break
		}
	}
}

func init() {
	flag.StringVar(&serverIP, "IP", "127.0.0.1", "设置服务器IP地址 默认为127.0.0.1")
	flag.IntVar(&serverPort, "Port", 8888, "设置服务器端口 默认为8888")
}




func main() {
	
	flag.Parse()
	
	client := NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println(">>>>>>>>链接服务器失败...")
		return
	}
	fmt.Println(">>>>>>>链接服务器成功")

	//单独开启一个goroutine处理服务器的返回
	go client.DealResponse()

	//启动客户端程序
	client.Run()
}