package main

import (
	"cinx/zinx"
	"cinx/zinx/ziface"
	"fmt"
)

type PingRouter struct {
	zinx.BaseRouter
}

// Handle Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//type HelloZinxRouter struct {
//	zinx.BaseRouter
//}

//func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
//	fmt.Println("Call HelloZinxRouter Handle")
//	//先读取客户端的数据，再回写ping...ping...ping
//	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
//
//	err := request.GetConnection().SendMsg(1, []byte("Hello Zinx Router V0.6"))
//	if err != nil {
//		fmt.Println(err)
//	}
//}

func main() {
	s := zinx.NewServer()
	s.AddRouter(1, &PingRouter{})
	//s.AddRouter(1, &HelloZinxRouter{})
	s.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Println("SetOnConnStart is Called ... ")
	})
	s.Serve()
}
