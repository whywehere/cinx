package ziface

import "net"

// IConnection 定义连接接⼝口
type IConnection interface {
	// Start 启动连接，让当前连接开始⼯工作
	Start()
	// Stop 停⽌止连接，结束当前连接状态M
	Stop()
	// GetConnID 从当前连接获取原始的socket TCPConn GetTCPConnection() *net.TCPConn //获取当前连接ID
	GetConnID() uint32
	GetTCPConnection() *net.TCPConn //获取远程客户端地址信息 RemoteAddr() net.Addr
	SendMsg(uint32, []byte) error
	//直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgId uint32, data []byte) error //添加带缓冲发送消息接口
	//设置链接属性
	SetProperty(key string, value interface{})
	//获取链接属性
	GetProperty(key string) (interface{}, error)
	//移除链接属性
	RemoveProperty(key string)
}

// HandFunc 定义⼀一个统⼀一处理理链接业务的接⼝口
type HandFunc func(*net.TCPConn, []byte, int) error
