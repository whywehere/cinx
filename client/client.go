package main

import (
	"cinx/zinx"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	fmt.Println("client start ... ")

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start failed: ", err)
		return
	}

	for {

		dp := zinx.NewDataPack()

		//将要发送的消息打包
		message := "Zinx V0.5 Client Test Message"
		msg, err := dp.Pack(zinx.NewMsgPackage(1, []byte(message)))
		if err != nil {
			return
		}
		_, err = conn.Write(msg)
		if err != nil {
			fmt.Println("write error: ", err)
			return
		}

		//接收消息
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData)
		if err != nil {
			fmt.Println("read buf error ")
			return
		}

		msgHead, err := dp.Unpack(headData)
		if err != nil {
			panic(err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			msg := msgHead.(*zinx.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		time.Sleep(1 * time.Second)
	}
}
